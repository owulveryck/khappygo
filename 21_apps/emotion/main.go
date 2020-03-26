package main

import (
	"context"
	"errors"
	"image"
	"image/jpeg"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"cloud.google.com/go/storage"
	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"
	"github.com/owulveryck/khappygo/common/emotions"
	"github.com/owulveryck/khappygo/common/kclient"
	"github.com/owulveryck/khappygo/common/machine"
	"gorgonia.org/tensor"
)

type configuration struct {
	Model string `envconfig:"model" required:"true"`
}

var (
	config        configuration
	storageClient *storage.Client
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
	storageClient, err = storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	rc, err := storageClient.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		log.Fatal(err)
	}
	kreceiver, err := kclient.NewDefaultClient()
	if err != nil {
		log.Fatal("Failed to create client, ", err)
	}
	machine := machine.NewModelMachine()
	err = machine.Start(rc)
	if err != nil {
		log.Fatal(err)
	}
	rc.Close()

	ep := &EventProcessor{
		Machine: machine,
	}

	log.Println("emotion is listening for events")
	log.Fatal(kreceiver.StartReceiver(context.Background(), ep.Receive))
}

// EventProcessor ...
type EventProcessor struct {
	Machine *machine.ModelMachine
}

// Receive ...
func (e *EventProcessor) Receive(ctx context.Context, event cloudevents.Event, response *cloudevents.EventResponse) error {
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

	jpg, _ := jpeg.Decode(rc)

	height := 64
	width := 64
	inputT := tensor.New(tensor.WithShape(1, 1, height, width), tensor.Of(tensor.Float32))

	m := imaging.Resize(jpg, height+30, width+30, imaging.Lanczos)

	// Crop the original image to 300x300px size using the center anchor.
	m = imaging.CropAnchor(m, height, width, imaging.Center)

	// Create a grayscale version of the image with higher contrast and sharpness.
	m = imaging.Grayscale(m)
	m = imaging.AdjustContrast(m, 20)
	//m = imaging.Sharpen(m, 2)

	var imgGray *image.Gray
	gray := imaging.Grayscale(m)
	imgGray = image.NewGray(gray.Bounds())
	for i := 0; i < len(imgGray.Pix); i++ {
		imgGray.Pix[i] = gray.Pix[i*4]
	}
	err = GrayToBCHW(imgGray, inputT)
	if err != nil {
		log.Println(err)
		response.Error(http.StatusInternalServerError, err.Error())
		return err
	}

	job := machine.NewJob(inputT)
	defer close(job.Output)
	defer close(job.ErrC)
	e.Machine.Feed <- job

	var outputs []tensor.Tensor
	select {
	case err := <-job.ErrC:
		log.Println(err)
		response.Error(http.StatusInternalServerError, err.Error())
		return err
	case outputs = <-job.Output:
	}
	emotionT := outputs[0].Data().([]float32)
	log.Println(emotionT)
	emotions := emotions.Emotion{
		Src:       imgPath,
		Neutral:   emotionT[0],
		Happiness: emotionT[1],
		Surprise:  emotionT[2],
		Sadness:   emotionT[3],
		Anger:     emotionT[4],
		Disgust:   emotionT[5],
		Fear:      emotionT[6],
		Contempt:  emotionT[7],
	}

	log.Printf("%#v", emotions)
	newEvent := cloudevents.NewEvent("1.0")
	newEvent.SetID(uuid.New().String())
	newEvent.SetSource("emotion")
	newEvent.SetDataContentType("application/json")
	newEvent.SetType("emotion")
	corrID, err := event.Context.GetExtension("correlation")
	if err != nil {
		newEvent.SetExtension("correlation", corrID)
	}
	newEvent.SetData(emotions)
	response.RespondWith(200, &newEvent)

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
