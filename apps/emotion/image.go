package main

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"reflect"

	"gorgonia.org/tensor"
	"gorgonia.org/tensor/native"
)

// GrayToBCHW convert an image to a BCHW tensor
// this function returns an error if:
//
//   - dst is not a pointer
//   - dst's shape is not 4
//   - dst' second dimension is not 1
//   - dst's third dimension != i.Bounds().Dy()
//   - dst's fourth dimension != i.Bounds().Dx()
//   - dst's type is not float32 or float64 (temporary)
func GrayToBCHW(img *image.Gray, dst tensor.Tensor) error {
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()
	err := verifyBCHWTensor(dst, h, w, true)
	if err != nil {
		return err
	}

	switch dst.Dtype() {
	case tensor.Float32:
		for x := 0; x < w; x++ {
			for y := 0; y < h; y++ {
				color := img.GrayAt(x, y)
				err := dst.SetAt(float32(color.Y)/255, 0, 0, y, x)
				if err != nil {
					return err
				}
			}
		}
	case tensor.Float64:
		for x := 0; x < w; x++ {
			for y := 0; y < h; y++ {
				color := img.GrayAt(x, y)
				err := dst.SetAt(float64(color.Y)/255, x, y)
				if err != nil {
					return err
				}
			}
		}
	default:
		return fmt.Errorf("%v not handled yet", dst.Dtype())
	}
	return nil
}

// TensorToImg turn a BCHW tensor into an image (BCHW with B=1)
func TensorToImg(t tensor.Tensor) (image.Image, error) {
	type img interface {
		image.Image
		Set(x, y int, c color.Color)
	}
	var output img
	if len(t.Shape()) != 4 {
		return nil, errors.New("expected a BCHW")
	}
	if t.Shape()[0] != 1 {
		return nil, errors.New("unhandled tensor with batch size > 1")
	}
	s := t.Shape()
	c, h, w := s[1], s[2], s[3]
	var rect = image.Rect(0, 0, w, h)
	t3, err := toTensor3(t)
	if err != nil {
		return nil, err
	}
	switch c {
	case 1:
		output = image.NewGray(rect)
	case 3:
		output = image.NewNRGBA(rect)
	default:
		return nil, errors.New("unhandled image encoding")
	}

	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			color, err := t3.getColor(y, x)
			if err != nil {
				return nil, err
			}
			output.Set(x, y, color)
		}
	}
	return output, nil
}

func verifyBCHWTensor(dst tensor.Tensor, h, w int, cowardMode bool) error {
	// check if tensor is a pointer
	rv := reflect.ValueOf(dst)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("cannot decode image into a non pointer or a nil receiver")
	}
	// check if tensor is compatible with BCHW (4 dimensions)
	if len(dst.Shape()) != 4 {
		return fmt.Errorf("Expected a 4 dimension tensor, but receiver has only %v", len(dst.Shape()))
	}
	// Check the batch size
	if dst.Shape()[0] != 1 {
		return errors.New("only batch size of one is supported")
	}
	if cowardMode && dst.Shape()[1] != 1 {
		return errors.New("Cowardly refusing to insert a gray scale into a tensor with more than one channel")
	}
	if dst.Shape()[2] != h || dst.Shape()[3] != w {
		return fmt.Errorf("cannot fit image into tensor; image is %v*%v but tensor is %v*%v", h, w, dst.Shape()[2], dst.Shape()[3])
	}
	return nil
}

// dumb structure to avoid type asertion at runtime
type tensor3 struct {
	c   int
	h   int
	w   int
	f32 [][][]float32
	f64 [][][]float64
	i32 [][][]int32
	i64 [][][]int64
}

func toTensor3(t tensor.Tensor) (*tensor3, error) {
	if len(t.Shape()) != 4 {
		return nil, errors.New("TensorToImg: expected a 4D tensor (BCHW)")
	}
	if t.Shape()[0] != 1 {
		return nil, errors.New("batch >1 not implemented")
	}
	dense, ok := t.(*tensor.Dense)
	if !ok {
		return nil, errors.New("This function can only convert dense tensors")
	}
	originalShape := make([]int, 4)
	newShape := make([]int, 3)
	copy(originalShape, t.Shape())
	copy(newShape, t.Shape()[1:4])
	err := dense.Reshape(newShape...)
	if err != nil {
		return nil, err
	}
	defer func() {
		dense.Reshape(originalShape...)
	}()

	if f32, err := native.Tensor3F32(dense); err == nil {
		return &tensor3{
			c:   dense.Shape()[0],
			h:   dense.Shape()[1],
			w:   dense.Shape()[2],
			f32: f32,
		}, nil
	}
	if f64, err := native.Tensor3F64(dense); err == nil {
		return &tensor3{
			c:   dense.Shape()[0],
			h:   dense.Shape()[1],
			w:   dense.Shape()[2],
			f64: f64,
		}, nil
	}

	if i32, err := native.Tensor3I32(dense); err == nil {
		return &tensor3{
			c:   dense.Shape()[0],
			h:   dense.Shape()[1],
			w:   dense.Shape()[2],
			i32: i32,
		}, nil
	}

	if i64, err := native.Tensor3I64(dense); err == nil {
		return &tensor3{
			c:   dense.Shape()[0],
			h:   dense.Shape()[1],
			w:   dense.Shape()[2],
			i64: i64,
		}, nil
	}

	return nil, errors.New("cannot convert to tensor3")
}
func (t *tensor3) getUint8(c, h, w int) (uint8, error) {
	lc := t.c
	lh := t.h
	lw := t.w
	if c > lc || h > lh || w > lw {
		return 0, errors.New("request out of bound")
	}
	switch {
	case t.f32 != nil:
		return uint8(t.f32[c][h][w]), nil
	case t.f64 != nil:
		return uint8(t.f64[c][h][w]), nil
	case t.i32 != nil:
		return uint8(t.i32[c][h][w]), nil
	case t.i64 != nil:
		return uint8(t.i64[c][h][w]), nil
	}
	return 0, nil
}

func (t *tensor3) getUint16(c, h, w int) (uint16, error) {
	lc := t.c
	lh := t.h
	lw := t.w
	if c > lc || h > lh || w > lw {
		return 0, errors.New("request out of bound")
	}
	switch {
	case t.f32 != nil:
		return uint16(t.f32[c][h][w]), nil
	case t.f64 != nil:
		return uint16(t.f64[c][h][w]), nil
	case t.i32 != nil:
		return uint16(t.i32[c][h][w]), nil
	case t.i64 != nil:
		return uint16(t.i64[c][h][w]), nil
	}
	return 0, nil
}

func (t *tensor3) getColor(h, w int) (color.Color, error) {
	switch t.c {
	case 1:
		y, err := t.getUint8(0, h, w)
		return color.Gray{
			Y: y,
		}, err
	case 3:
		r, err := t.getUint8(0, h, w)
		if err != nil {
			return nil, err
		}
		g, err := t.getUint8(1, h, w)
		if err != nil {
			return nil, err
		}
		b, err := t.getUint8(2, h, w)
		if err != nil {
			return nil, err
		}
		return color.NRGBA{
			R: r,
			G: g,
			B: b,
			A: uint8(255),
		}, nil
	default:
		return nil, errors.New("unhandled number of channel")
	}
}
