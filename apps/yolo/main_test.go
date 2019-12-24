package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/google/uuid"
	"github.com/owulveryck/onnx-go"
	"github.com/owulveryck/onnx-go/backend/x/gorgonnx"
)

func TestReceive(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r)
	}))
	defer ts.Close()
	b, err := ioutil.ReadFile("../../models/faces.onnx")
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
	kreceiver, err := newDefaultClient()
	if err != nil {
		t.Fatal("Failed to create client, ", err)
	}
	kclient, err := newDefaultClient(ts.URL)
	if err != nil {
		t.Fatal("Failed to create client, ", err)
	}
	c := &carrier{
		cloudeventsClient: kclient,
		model:             m,
		backend:           backend,
	}
	go func() {
		kreceiver.StartReceiver(context.Background(), c.receive)
	}()
	time.Sleep(5 * time.Second)
	newEvent := cloudevents.NewEvent()
	newEvent.SetSource("test")
	newEvent.SetType("image.png")
	newEvent.SetID(uuid.New().String())
	newEvent.SetData("file://WomaninaCrowd_400.jpg")
	_, response, err := c.cloudeventsClient.Send(context.Background(), newEvent)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(response)

}
