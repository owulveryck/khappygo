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
	"strings"

	"cloud.google.com/go/storage"
	cloudevents "github.com/cloudevents/sdk-go"
	pigo "github.com/esimov/pigo/core"
	"github.com/kelseyhightower/envconfig"
	"github.com/owulveryck/khappygo/common/kclient"
)

type configuration struct {
	Angle        float64 `default:"0.0"`
	MinSize      int     `default:"20"`
	MaxSize      int     `default:"1000"`
	ShiftFactor  float64 `default:"0.1"`
	ScaleFactor  float64 `default:"1.1"`
	IOUThreshold float64 `default:"0.2"`
	CascadeFile  string  `envconfig:"cascade_file" required:"true"`
}

var (
	config        configuration
	storageClient *storage.Client
	fd            *faceDetector
)

func main() {
	err := envconfig.Process("", &config)
	if err != nil {
		log.Fatal(envconfig.Usage("", &config))
	}
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
	rc.Close()

	p := pigo.NewPigo()
	// Unpack the binary file. This will return the number of cascade trees,
	// the tree depth, the threshold and the prediction from tree's leaf nodes.
	classifier, err := p.Unpack(cascadeFile)
	if err != nil {
		log.Fatal(err)
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

	kreceiver, err := kclient.NewDefaultClient()
	if err != nil {
		log.Fatal("Failed to create client, ", err)
	}
	log.Println("pigo is listening for events")

	log.Fatal(kreceiver.StartReceiver(context.Background(), receive))
}

func receive(ctx context.Context, event cloudevents.Event, response *cloudevents.EventResponse) error {
	log.Println(event)
	var imgPath string
	err := event.DataAs(&imgPath)
	if err != nil {
		log.Println(err)
		response.Error(http.StatusBadRequest, "expected data to be a string")
		return errors.New("expected data to be a string")
	}
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

	_, rects, _ := fd.drawFaces(faces, false)

	log.Println(detectionResult{
		coords: rects,
	})

	response.RespondWith(http.StatusOK, nil)
	return nil
}

func getElement(ctx context.Context, imgPath string) (io.ReadCloser, error) {
	imgPath = strings.Trim(imgPath, `"`)
	imageURL, err := url.Parse(imgPath)
	if err != nil {
		return nil, err
	}
	switch imageURL.Scheme {
	case "gs":
		bucket := imageURL.Host
		object := strings.Trim(imageURL.Path, "/")
		return storageClient.Bucket(bucket).Object(object).NewReader(ctx)
	case "file":
		return os.Open(imageURL.Host + imageURL.Path)
	}
	return nil, nil
}
