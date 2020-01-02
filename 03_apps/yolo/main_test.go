package main

import (
	"context"
	"io/ioutil"

	ceventHTTP "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"

	"net/http"
	"net/http/httptest"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/google/uuid"
	"github.com/owulveryck/khappygo/apps/common/box"
	"github.com/owulveryck/khappygo/apps/common/kclient"
	"github.com/owulveryck/onnx-go"
	"github.com/owulveryck/onnx-go/backend/x/gorgonnx"
)

func TestReceive(t *testing.T) {
	config = configuration{
		ConfidenceThreshold: 0.001,
		ClassProbaThreshold: 0.50,
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		msg, err := ceventHTTP.NewMessage(r.Header, r.Body)
		if err != nil {
			t.Fatal(err)
		}
		cdc := &ceventHTTP.Codec{
			Encoding: ceventHTTP.BinaryV1,
		}
		event, err := cdc.Decode(r.Context(), msg)
		if err != nil {
			t.Fatal(err)
		}
		var b box.Box
		err = event.DataAs(&b)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(b)

	}))
	defer ts.Close()
	//b, err := ioutil.ReadFile("../../models/tinyyolov2.onnx")
	b, err := ioutil.ReadFile("../../02_models/faces.onnx")
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
	kclient, err := kclient.NewDefaultClient(ts.URL)
	if err != nil {
		t.Fatal("Failed to create client, ", err)
	}
	c := &carrier{
		cloudeventsClient: kclient,
		model:             m,
		backend:           backend,
	}
	newEvent := cloudevents.NewEvent()
	newEvent.SetSource("test")
	newEvent.SetType("image.png")
	newEvent.SetID(uuid.New().String())
	//newEvent.SetData("file://../testdata/meme.jpg")
	newEvent.SetData("file://../testdata/100k-ai-faces-1.jpg")
	var response cloudevents.EventResponse
	err = c.receive(context.TODO(), newEvent, &response)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(response)

}
