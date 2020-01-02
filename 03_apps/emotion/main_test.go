package main

import (
	"context"
	"io/ioutil"
	"log"

	"testing"

	"cloud.google.com/go/storage"
	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/google/uuid"
	"github.com/owulveryck/onnx-go"
	"github.com/owulveryck/onnx-go/backend/x/gorgonnx"
)

func TestReceive(t *testing.T) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	b, err := ioutil.ReadFile("../../02_models/emotions.onnx")
	if err != nil {
		t.Fatal(err)
	}
	// Create a backend receiver
	backend := gorgonnx.NewGraph()
	// Create a model and set the execution backend
	m := onnx.NewModel(backend)
	// Decode it into the model
	err = m.UnmarshalBinary(b)
	if err != nil {
		t.Fatal(err)
	}

	c := &carrier{
		storageClient: client,
		model:         m,
		backend:       backend,
	}

	newEvent := cloudevents.NewEvent()
	newEvent.SetSource("image-extractor")
	newEvent.SetType("image.partial.png")
	newEvent.SetID(uuid.New().String())
	newEvent.SetData("file://../testdata/meme_0_face.jpg")
	var response cloudevents.EventResponse
	err = c.receive(context.TODO(), newEvent, &response)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(response)
}
