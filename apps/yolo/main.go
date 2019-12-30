package main

import (
	"bytes"
	"context"
	"errors"
	"image/jpeg"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"cloud.google.com/go/storage"
	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"
	"github.com/owulveryck/gofaces"
	"github.com/owulveryck/khappygo/apps/internal/box"
	"github.com/owulveryck/khappygo/apps/internal/kclient"
	"github.com/owulveryck/onnx-go"
	"github.com/owulveryck/onnx-go/backend"
	"github.com/owulveryck/onnx-go/backend/x/gorgonnx"
	"gorgonia.org/tensor"
)

type configuration struct {
	ConfidenceThreshold float64 `envconfig:"confidence_threshold" default:"0.10" required:"true"`
	ClassProbaThreshold float64 `envconfig:"proba_threshold" default:"0.90" required:"true"`
	Model               string  `envconfig:"model" required:"true"`
	Broker              string  `envconfig:"broker" required:"true"`
}

var (
	config configuration
)

func main() {
	err := envconfig.Process("", &config)
	if err != nil {
		log.Fatal(envconfig.Usage("", &config))
	}
	modelURL, err := url.Parse(config.Model)
	if err != nil {
		log.Fatal(err)
	}
	if modelURL.Scheme != "gs" {
		log.Fatal("Only model stored on Google Storage are supported")
	}
	bucket := modelURL.Host
	object := strings.Trim(modelURL.Path, "/")

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	rc, err := client.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Reading model")
	b, err := ioutil.ReadAll(rc)
	rc.Close()
	if err != nil {
		log.Fatal(err)
	}
	// Create a backend receiver
	backend := gorgonnx.NewGraph()
	// Create a model and set the execution backend
	m := onnx.NewModel(backend)
	// Decode it into the model
	log.Println("Unmarshaling model")
	err = m.UnmarshalBinary(b)
	if err != nil {
		log.Fatal(err)
	}
	kreceiver, err := kclient.NewDefaultClient()
	if err != nil {
		log.Fatal("Failed to create client, ", err)
	}
	kclient, err := kclient.NewDefaultClient(config.Broker)
	if err != nil {
		log.Fatal("Failed to create client, ", err)
	}
	c := &carrier{
		cloudeventsClient: kclient,
		storageClient:     client,
		model:             m,
		backend:           backend,
	}
	log.Println("save is listening for events")
	log.Fatal(kreceiver.StartReceiver(context.Background(), c.receive))
}

type carrier struct {
	cloudeventsClient cloudevents.Client
	storageClient     *storage.Client
	model             *onnx.Model
	backend           backend.ComputationBackend
}

func (c *carrier) receive(ctx context.Context, event cloudevents.Event, response *cloudevents.EventResponse) error {
	var imgPath string
	err := event.DataAs(&imgPath)
	if err != nil {
		log.Println(err)
		response.Error(http.StatusBadRequest, "expected data to be a string")
		return errors.New("expected data to be a string")
	}
	rc, err := c.getElement(ctx, imgPath)
	if err != nil {
		log.Println(err)
		response.Error(http.StatusBadRequest, err.Error())
		return err
	}
	defer rc.Close()
	var buf bytes.Buffer
	tee := io.TeeReader(rc, &buf)
	jpg, _ := jpeg.Decode(tee)
	log.Println(jpg.Bounds())

	inputT, err := gofaces.GetTensorFromImage(&buf)
	if err != nil {
		log.Println(err)
		response.Error(http.StatusInternalServerError, err.Error())
		return err
	}
	ratioX := float64(jpg.Bounds().Max.X) / float64(inputT.Shape()[1])
	ratioY := float64(jpg.Bounds().Max.Y) / float64(inputT.Shape()[2])
	ratio := ratioY
	if ratioX > ratioY {
		ratio = ratioX
	}
	c.model.SetInput(0, inputT)
	err = c.backend.Run()
	if err != nil {
		log.Println(err)
		response.Error(http.StatusInternalServerError, err.Error())
		return err
	}
	outputs, err := c.model.GetOutputTensors()

	boxes, err := gofaces.ProcessOutput(outputs[0].(*tensor.Dense))
	if err != nil {
		log.Println(err)
		response.Error(http.StatusInternalServerError, err.Error())
		return err
	}
	boxes = gofaces.Sanitize(boxes)
	if err != nil {
		log.Println(err)
		response.Error(http.StatusInternalServerError, err.Error())
		return err
	}

	output := make([]box.Box, 0)
	for i := 0; i < len(boxes); i++ {
		if boxes[i].Confidence >= config.ConfidenceThreshold {
			output = append(output, box.Box{
				Src:        imgPath,
				ID:         i,
				Element:    boxes[i].Elements[0].Class,
				Confidence: boxes[i].Confidence,
				X0:         int(float64(boxes[i].R.Min.X) * ratio),
				Y0:         int(float64(boxes[i].R.Min.Y) * ratio),
				X1:         int(float64(boxes[i].R.Max.X) * ratio),
				Y1:         int(float64(boxes[i].R.Max.Y) * ratio),
			})
		}
	}
	for i := 0; i < len(output); i++ {
		element := output[i].Element
		//		for _, element := range output[i].Elements {
		newEvent := cloudevents.NewEvent("1.0")
		newEvent.Context = event.Context.Clone()
		newEvent.SetType("boundingbox")
		newEvent.SetID(uuid.New().String())
		newEvent.SetSource("yolo")
		newEvent.SetData(output[i])
		newEvent.SetExtension("element", element)
		_, _, err = c.cloudeventsClient.Send(ctx, newEvent)
		if err != nil {
			log.Println(err)
			response.Error(http.StatusInternalServerError, err.Error())
			return err
		}
		//		}
	}
	response.RespondWith(http.StatusOK, nil)
	return nil
}

func (c *carrier) getElement(ctx context.Context, imgPath string) (io.ReadCloser, error) {
	imageURL, err := url.Parse(imgPath)
	if err != nil {
		return nil, err
	}
	switch imageURL.Scheme {
	case "gs":
		bucket := imageURL.Host
		object := strings.Trim(imageURL.Path, "/")
		return c.storageClient.Bucket(bucket).Object(object).NewReader(ctx)
	case "file":
		return os.Open(imageURL.Host + imageURL.Path)
	}
	return nil, nil
}
