package main

import (
	"context"
	"errors"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/owulveryck/khappygo/common/emotions"
	"github.com/owulveryck/khappygo/common/kclient"
)

var (
	client *firestore.Client
)

func main() {
	ctx := context.Background()
	var err error
	client, err = firestore.NewClient(ctx, "khappygo")
	if err != nil {
		log.Fatal(err)

	}
	kreceiver, err := kclient.NewDefaultClient()
	if err != nil {
		log.Fatal("Failed to create client, ", err)
	}
	log.Fatal(kreceiver.StartReceiver(context.Background(), Receive))
}

// Receive ...
func Receive(ctx context.Context, event cloudevents.Event, response *cloudevents.EventResponse) error {
	var emotions emotions.Emotion
	log.Println(event)
	log.Println(event.Data)
	err := event.DataAs(&emotions)
	if err != nil {
		log.Println(err)
		response.Error(http.StatusBadRequest, "expected data to be an emotion ")
		return errors.New("expected data to be an emotion")
	}
	_, _, err = client.Collection("khappygo").Add(ctx, emotions)
	if err != nil {
		log.Println(err)
		response.RespondWith(http.StatusInternalServerError, nil)
		return err
	}
	response.RespondWith(200, nil)
	return nil
}
