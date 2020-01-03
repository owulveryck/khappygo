package machine

import (
	"bytes"
	"io"
	"testing"

	onnxtest "github.com/owulveryck/onnx-go/backend/testbackend/onnx"
	"gorgonia.org/tensor"
)

func TestModelMachine_Start(t *testing.T) {
	// START_TEST OMIT
	machine := NewModelMachine()
	var onnxModelReader io.Reader
	// Reading onnxModelReader ... (ex: os.Open("path/to/model.onnx"))
	test := onnxtest.NewTestSqrt()                 // OMIT
	onnxModelReader = bytes.NewBuffer(test.ModelB) // OMIT
	err := machine.Start(onnxModelReader)
	if err != nil { //OMIT
		t.Fatal(err) //OMIT
	} //OMIT
	var inputTensor tensor.Tensor
	// Setting value... (ex: inputTensor = ImageToTensor("image.jpg"))
	inputTensor = test.Input[0] // OMIT
	job := NewJob(inputTensor)
	machine.Feed <- job
	select {
	case err := <-job.ErrC:
		// Error handling
		t.Fatal(err) // OMIT
	case outputs := <-job.Output:
		// outputs is an array of tensors
		// process the output
		t.Log(outputs[0]) // OMIT
	}
	// END_TEST OMIT
}
