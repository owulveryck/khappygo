package main

import (
	"context"
	"log"
	"os"

	"testing"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/google/uuid"
	"github.com/owulveryck/khappygo/common/machine"
)

func TestReceive(t *testing.T) {
	ctx := context.Background()
	f, err := os.Open("../../20_models/emotions.onnx")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	machine := machine.NewModelMachine()
	err = machine.Start(f)
	if err != nil {
		log.Fatal(err)
	}

	ep := &EventProcessor{
		Machine: machine,
	}

	newEvent := cloudevents.NewEvent("1.0")
	newEvent.SetSource("image-extractor")
	newEvent.SetType("image.partial.png")
	newEvent.SetID(uuid.New().String())
	newEvent.SetData(`"file://../testdata/face.jpg"`)
	var response cloudevents.EventResponse
	err = ep.Receive(ctx, newEvent, &response)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(response)
}
