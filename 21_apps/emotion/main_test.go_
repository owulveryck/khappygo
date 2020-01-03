package main

import (
	"context"
	"io/ioutil"
	"log"

	"testing"

	"cloud.google.com/go/storage"
	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/google/uuid"
)

func TestReceive(t *testing.T) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	b, err := ioutil.ReadFile("../../20_models/emotions.onnx")
	if err != nil {
		t.Fatal(err)
	}

	c := &carrier{
		storageClient: client,
		onnx:          b,
	}

	newEvent := cloudevents.NewEvent()
	newEvent.SetSource("image-extractor")
	newEvent.SetType("image.partial.png")
	newEvent.SetID(uuid.New().String())
	newEvent.SetData(`"gs://khappygo/processed/meme_1_face.jpg"`)
	var response cloudevents.EventResponse
	err = c.receive(context.TODO(), newEvent, &response)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(response)
}
