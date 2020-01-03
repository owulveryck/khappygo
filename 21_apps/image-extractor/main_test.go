package main

import (
	"context"
	"log"

	"testing"

	"cloud.google.com/go/storage"
	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/google/uuid"
	"github.com/owulveryck/khappygo/common/box"
)

func TestReceive(t *testing.T) {
	config = configuration{
		Dest: "file:///tmp",
	}
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	c := &carrier{
		storageClient: client,
	}

	newEvent := cloudevents.NewEvent()
	newEvent.SetSource("test")
	newEvent.SetType("image.png")
	newEvent.SetID(uuid.New().String())
	/*
			    main_test.go:43: {0 file://../testdata/meme.jpg face 0.1957242087202336 187 85 251 147}
		    main_test.go:43: {1 file://../testdata/meme.jpg face 0.10400465095773953 0 126 104 206}
	*/
	/*
			    main_test.go:43: {0 file://../testdata/meme.jpg face 0.1957242087202336 348 158 468 274}
		    main_test.go:43: {1 file://../testdata/meme.jpg face 0.10400465095773953 0 235 194 384}
	*/
	for _, b := range []box.Box{
		box.Box{
			ID:         0,
			Src:        "file://../testdata/meme.jpg",
			Element:    "face",
			Confidence: 0.1957242087202336,
			X0:         348,
			Y0:         158,
			X1:         468,
			Y1:         274,
		},
		//    main_test.go:43: {9 file://../testdata/100k-ai-faces-1.jpg face 0.010132860146448295 706 364 750 385}
		box.Box{
			ID:         0,
			Src:        "file://../testdata/100k-ai-faces-1.jpg",
			Element:    "face",
			Confidence: 0.010132860146448295,
			X0:         705,
			Y0:         364,
			X1:         750,
			Y1:         385,
		},
	} {
		newEvent.SetData(b)
		var response cloudevents.EventResponse
		err = c.receive(context.TODO(), newEvent, &response)
		if err != nil {
			t.Fatal(err)
		}

		t.Log(response)

	}
}
