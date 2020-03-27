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
	"path/filepath"
	"strconv"
	"strings"

	"cloud.google.com/go/storage"
	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"
	"github.com/owulveryck/khappygo/common/box"
)

type configuration struct {
	Dest string `envconfig:"dest" required:"true"`
	Port int    `envconfig:"PORT" default:"8080"`
}

var (
	config configuration
)

func main() {

	err := envconfig.Process("", &config)
	if err != nil {
		log.Fatal(envconfig.Usage("", &config))
	}
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	t, err := cloudevents.NewHTTPTransport(
		cloudevents.WithPort(config.Port),
		cloudevents.WithPath("/"),
	)
	// or a custom transport: t := &custom.MyTransport{Cool:opts}

	kreceiver, err := cloudevents.NewClient(t)
	//	kreceiver, err := kclient.NewDefaultClient()
	if err != nil {
		log.Fatal("Failed to create client, ", err)
	}
	c := &carrier{
		storageClient: client,
	}
	log.Println("save is listening for events")
	log.Fatal(kreceiver.StartReceiver(context.Background(), c.receive))
}

type carrier struct {
	cloudeventsClient cloudevents.Client
	storageClient     *storage.Client
}

func (c *carrier) receive(ctx context.Context, event cloudevents.Event, response *cloudevents.EventResponse) error {
	var b box.Box
	err := event.DataAs(&b)
	if err != nil {
		log.Println(err)
		response.Error(http.StatusBadRequest, "expected data to be a string")
		return errors.New("expected data to be a string")
	}
	filename := filepath.Base(b.Src)
	var extension = filepath.Ext(filename)
	var name = filename[0 : len(filename)-len(extension)]

	rc, err := c.getElement(ctx, b.Src)
	if err != nil {
		log.Println(err)
		response.Error(http.StatusBadRequest, "expected data to be a string")
		return errors.New("expected data to be a string")
	}
	defer rc.Close()
	img, err := jpeg.Decode(rc)
	if err != nil {
		log.Println(err)
		response.Error(http.StatusInternalServerError, "cannot decode picture"+err.Error())
		return err
	}
	cropped := imaging.Crop(img, image.Rect(b.X0, b.Y0, b.X1, b.Y1))
	imgPath := config.Dest + "/" + name + "_" + strconv.Itoa(b.ID) + "_" + b.Element + extension
	w, err := c.postElement(ctx, imgPath)
	if err != nil {
		log.Println(err)
		response.Error(http.StatusInternalServerError, "save picture"+err.Error())
		return err
	}
	defer w.Close()
	err = jpeg.Encode(w, cropped, nil)
	if err != nil {
		log.Println(err)
		response.Error(http.StatusInternalServerError, "save picture"+err.Error())
		return err
	}

	newEvent := cloudevents.NewEvent()
	newEvent.SetID(uuid.New().String())
	newEvent.SetSource("image-extractor")
	newEvent.SetType("image.partial.png")
	corrID, err := event.Context.GetExtension("correlation")
	if err != nil {
		newEvent.SetExtension("correlation", corrID)
	}
	newEvent.SetExtension("element", b.Element)
	newEvent.SetData(imgPath)
	response.RespondWith(200, &newEvent)
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
func (c *carrier) postElement(ctx context.Context, imgPath string) (io.WriteCloser, error) {
	log.Println(imgPath)
	imageURL, err := url.Parse(imgPath)
	if err != nil {
		return nil, err
	}
	switch imageURL.Scheme {
	case "gs":
		bucket := imageURL.Host
		object := strings.Trim(imageURL.Path, "/")
		return c.storageClient.Bucket(bucket).Object(object).NewWriter(ctx), nil
	case "file":
		return os.Create(filepath.Join(imageURL.Host, imageURL.Path))
	}
	return nil, nil
}
