package machine

import (
	"bytes"
	"log"
	"testing"

	onnxtest "github.com/owulveryck/onnx-go/backend/testbackend/onnx"
)

func TestModelMachine_Start(t *testing.T) {
	machine := NewModelMachine()
	test := onnxtest.NewTestSqrt()
	err := machine.Start(bytes.NewBuffer(test.ModelB))
	if err != nil {
		t.Fatal(err)
	}
	job := NewJob(test.Input[0])
	machine.Feed <- job
	select {
	case err := <-job.ErrC:
		t.Fatal(err)
	case outputs := <-job.Output:
		log.Println("received")
		t.Log(outputs[0])
		//ExpectedOutput
	}
}
