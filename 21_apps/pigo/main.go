package main

import (
	"context"
	"errors"
	"image"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	gpubsub "cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
	cloudevents "github.com/cloudevents/sdk-go"

	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport/pubsub"
	pigo "github.com/esimov/pigo/core"
	"github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"
	"github.com/owulveryck/khappygo/common/box"
)

type configuration struct {
	Angle        float64 `default:"0.0"`
	MinSize      int     `default:"20"`
	MaxSize      int     `default:"1000"`
	ShiftFactor  float64 `default:"0.1"`
	ScaleFactor  float64 `default:"1.1"`
	IOUThreshold float64 `default:"0.01"`
	CascadeFile  string  `envconfig:"cascade_file" required:"true"`
	Port         int     `envconfig:"PORT" default:"8080"`
}

var (
	config        configuration
	storageClient *storage.Client
	fd            *faceDetector
	eventsClient  cloudevents.Client
)

func main() {
	err := envconfig.Process("", &config)
	if err != nil {
		log.Fatal(envconfig.Usage("", &config))
	}
	log.Printf("%#v", config)
	log.Println(config.CascadeFile)
	cascadeURL, err := url.Parse(config.CascadeFile)
	if err != nil {
		log.Fatal(err)
	}
	if cascadeURL.Scheme != "gs" {
		log.Fatal("Only model stored on Google Storage are supported")
	}
	bucket := cascadeURL.Host
	object := strings.Trim(cascadeURL.Path, "/")

	ctx := context.Background()
	storageClient, err = storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	rc, err := storageClient.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		log.Fatal(err)
	}
	cascadeFile, err := ioutil.ReadAll(rc)
	if err != nil {
		log.Fatal(err)
	}

	p := pigo.NewPigo()
	// Unpack the binary file. This will return the number of cascade trees,
	// the tree depth, the threshold and the prediction from tree's leaf nodes.
	classifier, err := p.Unpack(cascadeFile)
	rc.Close()
	if err != nil {
		log.Fatal(err)
	}
	pubsubClient, err := gpubsub.NewClient(ctx, "aerobic-botany-270918")
	if err != nil {
		log.Fatal(err)
	}

	tr, err := pubsub.New(ctx, pubsub.WithClient(pubsubClient), pubsub.WithTopicIDFromDefaultEnv())
	if err != nil {
		log.Fatal(err)
	}

	eventsClient, err = cloudevents.NewClient(tr)
	if err != nil {
		log.Fatal("Failed to create client, ", err)
	}

	fd = &faceDetector{
		angle:         config.Angle,
		classifier:    classifier,
		minSize:       config.MinSize,
		maxSize:       config.MaxSize,
		shiftFactor:   config.ShiftFactor,
		scaleFactor:   config.ScaleFactor,
		iouThreshold:  config.IOUThreshold,
		puploc:        false,
		puplocCascade: "",
		flploc:        false,
		flplocDir:     "",
		markDetEyes:   false,
	}

	t, err := cloudevents.NewHTTPTransport(
		cloudevents.WithPort(config.Port),
	)
	// or a custom transport: t := &custom.MyTransport{Cool:opts}

	kreceiver, err := cloudevents.NewClient(t)
	if err != nil {
		log.Fatal("Failed to create client, ", err)
	}
	log.Println("pigo is listening for events")

	log.Fatal(kreceiver.StartReceiver(context.Background(), receive))
}

func receive(ctx context.Context, event cloudevents.Event, response *cloudevents.EventResponse) error {
	log.Println("received event: " + event.Type() + " " + event.Source() + " " + event.Subject())
	var payload eventProtoPayload
	err := event.DataAs(&payload)
	if err != nil {
		log.Println(err)
		response.Error(http.StatusBadRequest, "expected data to be a ProtoPayload")
		return errors.New("expected data to be a ProtoPayload")
	}
	if payload.ProtoPayload.ServiceName != "storage.googleapis.com" &&
		payload.ProtoPayload.MethodName != "storage.objects.create" {
		return nil
	}
	if !strings.Contains(payload.ProtoPayload.ResourceName, "-input") {
		return nil
	}
	if filepath.Ext(payload.ProtoPayload.ResourceName) != ".jpg" {
		return nil
	}
	rplcr := strings.NewReplacer("projects/_/buckets/", "gs://", "objects/", "")
	imgPath := rplcr.Replace(payload.ProtoPayload.ResourceName)
	log.Println(imgPath)
	rc, err := getElement(ctx, imgPath)
	if err != nil {
		log.Println(err)
		response.Error(http.StatusBadRequest, err.Error())
		return err
	}
	defer rc.Close()

	src, _, err := image.Decode(rc)
	if err != nil {
		log.Println(err)
		response.Error(http.StatusBadRequest, err.Error())
		return err
	}
	img := pigo.ImgToNRGBA(src)
	faces, err := fd.detectFaces(img)
	if err != nil {
		log.Fatalf("Detection error: %v", err)
	}

	output := make([]box.Box, len(faces))
	var qThresh float32 = 5.0

	log.Printf("%#v", faces)
	for i, face := range faces {
		if face.Q > qThresh {
			output = append(output, box.Box{
				Src:        imgPath,
				ID:         i,
				Element:    "face",
				Confidence: float64(face.Q),
				X0:         int(float64(face.Col - face.Scale/2)),
				Y0:         int(float64(face.Row - face.Scale/2)),
				X1:         int(float64(face.Col + face.Scale/2)),
				Y1:         int(float64(face.Row + face.Scale/2)),
			})

		}
	}
	for i := 0; i < len(output); i++ {
		if output[i].Src == "" {
			continue
		}
		element := output[i].Element
		//		for _, element := range output[i].Elements {
		newEvent := cloudevents.NewEvent("1.0")
		//log.Println(event.Context)
		//newEvent.Context = event.Context.Clone()
		newEvent.SetType("boundingbox")
		newEvent.SetID(uuid.New().String())
		newEvent.SetSource("pigo")
		newEvent.SetExtension("correlation", uuid.New().String())
		newEvent.SetData(output[i])
		newEvent.SetExtension("element", element)
		log.Println("Sending event: ", newEvent)
		_, _, err = eventsClient.Send(ctx, newEvent)
		if err != nil {
			log.Println(err)
			response.Error(http.StatusInternalServerError, err.Error())
			return err
		}
	}

	response.RespondWith(http.StatusOK, nil)
	return nil
}

func getElement(ctx context.Context, imgPath string) (io.ReadCloser, error) {
	log.Println(imgPath)
	imgPath = strings.Trim(imgPath, `"`)
	imageURL, err := url.Parse(imgPath)
	if err != nil {
		return nil, err
	}
	switch imageURL.Scheme {
	case "gs":
		bucket := imageURL.Host
		if filepath.Ext(imageURL.Path) != ".jpg" {
			return nil, errors.New("not a jpg file:" + imageURL.Path)
		}
		object := strings.Trim(imageURL.Path, "/")
		return storageClient.Bucket(bucket).Object(object).NewReader(ctx)
	case "file":
		return os.Open(imageURL.Host + imageURL.Path)
	}
	return nil, errors.New("unsupported sheme")
}
