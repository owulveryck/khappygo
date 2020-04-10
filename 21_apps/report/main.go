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
	client, err = firestore.NewClient(ctx, "aerobic-botany-270918")
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
	data, err := event.DataBytes()
	if err != nil {
		log.Println(err)
		response.Error(http.StatusBadRequest, err.Error())
		return err
	}
	var emotions emotions.Emotion
	err = unmarshalData(data, &emotions)
	if err != nil {
		log.Println(err)
		response.Error(http.StatusBadRequest, "expected data to be a emotion")
		return errors.New("expected data to be a emotion")
	}
	log.Println("Storing: ", emotions)
	_, _, err = client.Collection("khappygo").Add(ctx, emotions)
	if err != nil {
		log.Println(err)
		response.RespondWith(http.StatusInternalServerError, nil)
		return err
	}
	response.RespondWith(200, nil)
	return nil
}
