package main

import (
	"context"
	"io"
	"log"
	"net/url"
	"os"
	"strings"

	"cloud.google.com/go/storage"
	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/kelseyhightower/envconfig"
	"github.com/owulveryck/khappygo/apps/internal/kclient"
)

type configuration struct {
	Broker string `envconfig:"broker" required:"true"`
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
	}
	log.Println("save is listening for events")
	log.Fatal(kreceiver.StartReceiver(context.Background(), c.receive))
}

type carrier struct {
	cloudeventsClient cloudevents.Client
	storageClient     *storage.Client
}

func (c *carrier) receive(ctx context.Context, event cloudevents.Event, response *cloudevents.EventResponse) error {
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
