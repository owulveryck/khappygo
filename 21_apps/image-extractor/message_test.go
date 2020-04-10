package main

import (
	"testing"

	"github.com/owulveryck/khappygo/common/box"
)

func Test_unmarshalData(t *testing.T) {
	testData := []byte(`{"message":{"attributes":{"ce-correlation":"ff4aba76-80a5-4cce-a5df-ec74dee33ded","ce-datacontenttype":"application/json","ce-element":"face","ce-id":"e5275280-b906-4ddf-b29f-536bc6f4869d","ce-source":"pigo","ce-specversion":"1.0","ce-type":"boundingbox"},"data":"eyJJRCI6OSwiU3JjIjoiZ3M6Ly9hZXJvYmljLWJvdGFueS0yNzA5MTgtaW5wdXQvdGVzdDI4LmpwZyIsIkVsZW1lbnQiOiJmYWNlIiwiQ29uZmlkZW5jZSI6MjQ2LjA1ODUxNzQ1NjA1NDcsIlgwIjo1MTAsIlkwIjoyNTYsIlgxIjo1ODYsIlkxIjozMzJ9","messageId":"1096207710802819","message_id":"1096207710802819","publishTime":"2020-04-07T15:43:24.048Z","publish_time":"2020-04-07T15:43:24.048Z"},"subscription":"projects/aerobic-botany-270918/subscriptions/cre-us-central1-image-extractor-sub-000"}`)
	var b box.Box
	err := unmarshalData(testData, &b)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(b)
}
