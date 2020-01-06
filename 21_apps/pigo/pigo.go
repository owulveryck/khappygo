package main

import (
	"image"
	"io/ioutil"

	pigo "github.com/esimov/pigo/core"
	"github.com/fogleman/gg"
)

const banner = `
┌─┐┬┌─┐┌─┐
├─┘││ ┬│ │
┴  ┴└─┘└─┘
Go (Golang) Face detection library.
    Version: %s
`

// Version indicates the current build version.
var Version string

var (
	dc        *gg.Context
	plc       *pigo.PuplocCascade
	flpcs     map[string][]*pigo.FlpCascade
	imgParams *pigo.ImageParams
)

var (
	eyeCascades  = []string{"lp46", "lp44", "lp42", "lp38", "lp312"}
	mouthCascade = []string{"lp93", "lp84", "lp82", "lp81"}
)

// faceDetector struct contains Pigo face detector general settings.
type faceDetector struct {
	angle         float64
	classifier    *pigo.Pigo
	destination   string
	minSize       int
	maxSize       int
	shiftFactor   float64
	scaleFactor   float64
	iouThreshold  float64
	puploc        bool
	puplocCascade string
	flploc        bool
	flplocDir     string
	markDetEyes   bool
}

// detectionResult contains the coordinates of the detected faces and the base64 converted image.
type detectionResult struct {
	coords []image.Rectangle
}

// detectFaces run the detection algorithm over the provided source image.
func (fd *faceDetector) detectFaces(src *image.NRGBA) ([]pigo.Detection, error) {

	pixels := pigo.RgbToGrayscale(src)
	cols, rows := src.Bounds().Max.X, src.Bounds().Max.Y

	dc = gg.NewContext(cols, rows)
	dc.DrawImage(src, 0, 0)

	imgParams = &pigo.ImageParams{
		Pixels: pixels,
		Rows:   rows,
		Cols:   cols,
		Dim:    cols,
	}

	cParams := pigo.CascadeParams{
		MinSize:     fd.minSize,
		MaxSize:     fd.maxSize,
		ShiftFactor: fd.shiftFactor,
		ScaleFactor: fd.scaleFactor,
		ImageParams: *imgParams,
	}

	if fd.puploc {
		pl := pigo.NewPuplocCascade()

		cascade, err := ioutil.ReadFile(fd.puplocCascade)
		if err != nil {
			return nil, err
		}
		plc, err = pl.UnpackCascade(cascade)
		if err != nil {
			return nil, err
		}

		if fd.flploc {
			flpcs, err = pl.ReadCascadeDir(fd.flplocDir)
			if err != nil {
				return nil, err
			}
		}
	}

	// Run the classifier over the obtained leaf nodes and return the detection results.
	// The result contains quadruplets representing the row, column, scale and detection score.
	faces := fd.classifier.RunCascade(cParams, fd.angle)

	// Calculate the intersection over union (IoU) of two clusters.
	faces = fd.classifier.ClusterDetections(faces, fd.iouThreshold)

	return faces, nil
}

// inSlice checks if the item exists in the slice.
func inSlice(item string, slice []string) bool {
	for _, it := range slice {
		if it == item {
			return true
		}
	}
	return false
}
