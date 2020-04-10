package main

import (
	"encoding/base64"
	"encoding/json"
	"time"
)

type message struct {
	Message struct {
		Attributes struct {
			CeCorrelation     string `json:"ce-correlation"`
			CeDatacontenttype string `json:"ce-datacontenttype"`
			CeElement         string `json:"ce-element"`
			CeID              string `json:"ce-id"`
			CeSource          string `json:"ce-source"`
			CeSpecversion     string `json:"ce-specversion"`
			CeType            string `json:"ce-type"`
		} `json:"attributes"`
		Data        string    `json:"data"`
		MessageID   string    `json:"message_id"`
		PublishTime time.Time `json:"publish_time"`
	} `json:"message"`
	Subscription string `json:"subscription"`
}

func unmarshalData(b []byte, i interface{}) error {
	var m message
	err := json.Unmarshal(b, &m)
	if err != nil {
		return err
	}
	data, err := base64.StdEncoding.DecodeString(m.Message.Data)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, i)
	return err
}
