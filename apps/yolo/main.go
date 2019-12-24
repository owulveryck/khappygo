package main

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
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
	"github.com/owulveryck/onnx-go"
	"github.com/owulveryck/onnx-go/backend"
	"github.com/owulveryck/onnx-go/backend/x/gorgonnx"
	"gorgonia.org/tensor"
)

type configuration struct {
	ConfidenceThreshold float64 `envconfig:"confidence_threshold" default:"0.10" required:"true"`
	ClassProbaThreshold float64 `envconfig:"proba_threshold" default:"0.90" required:"true"`
	Model               string  `envconfig:"model" required:"true"`
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
	kclient, err := newDefaultClient()
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
	log.Fatal(kclient.StartReceiver(context.Background(), c.receive))
}

type carrier struct {
	cloudeventsClient cloudevents.Client
	storageClient     *storage.Client
	model             *onnx.Model
	backend           backend.ComputationBackend
}

func (c *carrier) receive(ctx context.Context, event cloudevents.Event, response *cloudevents.EventResponse) error {
	log.Println(event.String())
	var imgPath string
	err := event.DataAs(&imgPath)
	if err != nil {
		log.Println(err)
		response.Error(http.StatusBadRequest, "expected data to be a string")
		return errors.New("expected data to be a string")
	}
	imageURL, err := url.Parse(imgPath)
	if err != nil {
		log.Println(err)
		response.Error(http.StatusBadRequest, err.Error())
		return err
	}
	if imageURL.Scheme != "gs" {
		response.Error(http.StatusBadRequest, "Only model stored on Google Storage are supported")
		return errors.New("Only model stored on Google Storage are supported")
	}
	bucket := imageURL.Host
	object := strings.Trim(imageURL.Path, "/")
	rc, err := c.storageClient.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		log.Println(err)
		response.Error(http.StatusBadRequest, err.Error())
		return err
	}
	defer rc.Close()

	inputT, err := gofaces.GetTensorFromImage(rc)
	if err != nil {
		log.Println(err)
		response.Error(http.StatusInternalServerError, err.Error())
		return err
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

	for i := 1; i < len(boxes); i++ {
		if boxes[i].Confidence < config.ConfidenceThreshold {
			boxes = boxes[:i]
			//continue
		}
	}
	for i := 1; i < len(boxes); i++ {
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		err := enc.Encode(imgPath)
		if err != nil {
			log.Println(err)
			response.Error(http.StatusInternalServerError, err.Error())
			return err
		}
		err = enc.Encode(boxes[i])
		if err != nil {
			log.Println(err)
			response.Error(http.StatusInternalServerError, err.Error())
			return err
		}
		data := buf.Bytes()
		for _, element := range boxes[i].Elements {
			newEvent := cloudevents.NewEvent()
			newEvent.Context = event.Context.Clone()
			newEvent.SetType("boundingbox")
			newEvent.SetID(uuid.New().String())
			newEvent.SetSource("yolo")
			newEvent.SetData(data)
			newEvent.SetExtension("element", element)
			_, _, err := c.cloudeventsClient.Send(ctx, newEvent)
			if err != nil {
				log.Println(err)
				response.Error(http.StatusInternalServerError, err.Error())
				return err
			}
		}
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
