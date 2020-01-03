package main

import (
	"io"
	"io/ioutil"
	"log"

	"github.com/owulveryck/onnx-go"
	"github.com/owulveryck/onnx-go/backend/x/gorgonnx"
	"gorgonia.org/tensor"
)

// NewModelMachine ...
func NewModelMachine() *ModelMachine {
	return &ModelMachine{
		Feed: make(chan Job, 0),
	}
}

type ModelMachine struct {
	Feed chan Job
}

type Job struct {
	InputT tensor.Tensor
	Output chan []tensor.Tensor
	ErrC   chan error
}

func NewJob(input tensor.Tensor) Job {
	return Job{
		InputT: input,
		Output: make(chan []tensor.Tensor, 0),
		ErrC:   make(chan error, 0),
	}
}

func (c *ModelMachine) Start(r io.Reader) error {
	log.Println("Reading model")
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	// Create a backend receiver
	backend := gorgonnx.NewGraph()
	// Create a model and set the execution backend
	model := onnx.NewModel(backend)
	// Decode it into the model
	log.Println("Unmarshaling model")
	err = model.UnmarshalBinary(b)
	if err != nil {
		return err
	}
	go func() {
		for job := range c.Feed {
			model.SetInput(0, job.InputT)
			err = backend.Run()
			if err != nil {
				job.ErrC <- err
				continue
			}
			outputs, err := model.GetOutputTensors()
			if err != nil {
				job.ErrC <- err
				continue
			}
			job.Output <- outputs
		}
	}()
	return nil
}
