package main

import (
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"math"
	"time"

	"github.com/disintegration/imaging"
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

// drawFaces marks the detected faces with a circle in case isCircle is true, otherwise marks with a rectangle.
func (fd *faceDetector) drawFaces(faces []pigo.Detection, isCircle bool) ([]byte, []image.Rectangle, error) {
	var (
		qThresh float32 = 5.0
		perturb         = 63
		rects   []image.Rectangle
		puploc  *pigo.Puploc
	)

	for _, face := range faces {
		if face.Q > qThresh {
			if isCircle {
				dc.DrawArc(
					float64(face.Col),
					float64(face.Row),
					float64(face.Scale/2),
					0,
					2*math.Pi,
				)
			} else {
				dc.DrawRectangle(
					float64(face.Col-face.Scale/2),
					float64(face.Row-face.Scale/2),
					float64(face.Scale),
					float64(face.Scale),
				)
			}
			rects = append(rects, image.Rect(
				face.Col-face.Scale/2,
				face.Row-face.Scale/2,
				face.Scale,
				face.Scale,
			))
			dc.SetLineWidth(2.0)
			dc.SetStrokeStyle(gg.NewSolidPattern(color.RGBA{R: 255, G: 0, B: 0, A: 255}))
			dc.Stroke()

			if fd.puploc && face.Scale > 50 {
				rect := image.Rect(
					face.Col-face.Scale/2,
					face.Row-face.Scale/2,
					face.Col+face.Scale/2,
					face.Row+face.Scale/2,
				)
				rows, cols := rect.Max.X-rect.Min.X, rect.Max.Y-rect.Min.Y
				ctx := gg.NewContext(rows, cols)
				faceZone := ctx.Image()

				// left eye
				puploc = &pigo.Puploc{
					Row:      face.Row - int(0.075*float32(face.Scale)),
					Col:      face.Col - int(0.175*float32(face.Scale)),
					Scale:    float32(face.Scale) * 0.25,
					Perturbs: perturb,
				}
				leftEye := plc.RunDetector(*puploc, *imgParams, fd.angle, false)
				if leftEye.Row > 0 && leftEye.Col > 0 {
					if fd.angle > 0 {
						drawDetections(ctx,
							float64(cols/2-(face.Col-leftEye.Col)),
							float64(rows/2-(face.Row-leftEye.Row)),
							float64(leftEye.Scale),
							color.RGBA{R: 255, G: 0, B: 0, A: 255},
							fd.markDetEyes,
						)
						angle := (fd.angle * 180) / math.Pi
						rotated := imaging.Rotate(faceZone, 2*angle, color.Transparent)
						final := imaging.FlipH(rotated)

						dc.DrawImage(final, face.Col-face.Scale/2, face.Row-face.Scale/2)
					} else {
						drawDetections(dc,
							float64(leftEye.Col),
							float64(leftEye.Row),
							float64(leftEye.Scale),
							color.RGBA{R: 255, G: 0, B: 0, A: 255},
							fd.markDetEyes,
						)
					}
				}

				// right eye
				puploc = &pigo.Puploc{
					Row:      face.Row - int(0.075*float32(face.Scale)),
					Col:      face.Col + int(0.185*float32(face.Scale)),
					Scale:    float32(face.Scale) * 0.25,
					Perturbs: perturb,
				}

				rightEye := plc.RunDetector(*puploc, *imgParams, fd.angle, false)
				if rightEye.Row > 0 && rightEye.Col > 0 {
					if fd.angle > 0 {
						drawDetections(ctx,
							float64(cols/2-(face.Col-rightEye.Col)),
							float64(rows/2-(face.Row-rightEye.Row)),
							float64(rightEye.Scale),
							color.RGBA{R: 255, G: 0, B: 0, A: 255},
							fd.markDetEyes,
						)
						// convert radians to angle
						angle := (fd.angle * 180) / math.Pi
						rotated := imaging.Rotate(faceZone, 2*angle, color.Transparent)
						final := imaging.FlipH(rotated)

						dc.DrawImage(final, face.Col-face.Scale/2, face.Row-face.Scale/2)
					} else {
						drawDetections(dc,
							float64(rightEye.Col),
							float64(rightEye.Row),
							float64(rightEye.Scale),
							color.RGBA{R: 255, G: 0, B: 0, A: 255},
							fd.markDetEyes,
						)
					}
				}
				if fd.flploc {
					for _, eye := range eyeCascades {
						for _, flpc := range flpcs[eye] {
							flp := flpc.FindLandmarkPoints(leftEye, rightEye, *imgParams, perturb, false)
							if flp.Row > 0 && flp.Col > 0 {
								drawDetections(dc,
									float64(flp.Col),
									float64(flp.Row),
									float64(flp.Scale*0.5),
									color.RGBA{R: 0, G: 0, B: 255, A: 255},
									false,
								)
							}

							flp = flpc.FindLandmarkPoints(leftEye, rightEye, *imgParams, perturb, true)
							if flp.Row > 0 && flp.Col > 0 {
								drawDetections(dc,
									float64(flp.Col),
									float64(flp.Row),
									float64(flp.Scale*0.5),
									color.RGBA{R: 0, G: 0, B: 255, A: 255},
									false,
								)
							}
						}
					}

					for _, mouth := range mouthCascade {
						for _, flpc := range flpcs[mouth] {
							flp := flpc.FindLandmarkPoints(leftEye, rightEye, *imgParams, perturb, false)
							if flp.Row > 0 && flp.Col > 0 {
								drawDetections(dc,
									float64(flp.Col),
									float64(flp.Row),
									float64(flp.Scale*0.5),
									color.RGBA{R: 0, G: 0, B: 255, A: 255},
									false,
								)
							}
						}
					}
					flp := flpcs["lp84"][0].FindLandmarkPoints(leftEye, rightEye, *imgParams, perturb, true)
					if flp.Row > 0 && flp.Col > 0 {
						drawDetections(dc,
							float64(flp.Col),
							float64(flp.Row),
							float64(flp.Scale*0.5),
							color.RGBA{R: 0, G: 0, B: 255, A: 255},
							false,
						)
					}
				}
			}
		}
	}

	return nil, rects, nil
}

type spinner struct {
	stopChan chan struct{}
}

// Start process
func (s *spinner) start(message string) {
	s.stopChan = make(chan struct{}, 1)

	go func() {
		for {
			for _, r := range `⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏` {
				select {
				case <-s.stopChan:
					return
				default:
					fmt.Printf("\r%s%s %c%s", message, "\x1b[35m", r, "\x1b[39m")
					time.Sleep(time.Millisecond * 100)
				}
			}
		}
	}()
}

// End process
func (s *spinner) stop() {
	s.stopChan <- struct{}{}
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

// drawDetections helper function to draw the detection marks
func drawDetections(ctx *gg.Context, x, y, r float64, c color.RGBA, markDet bool) {
	ctx.DrawArc(x, y, r*0.15, 0, 2*math.Pi)
	ctx.SetFillStyle(gg.NewSolidPattern(c))
	ctx.Fill()

	if markDet {
		ctx.DrawRectangle(x-(r*1.5), y-(r*1.5), r*3, r*3)
		ctx.SetLineWidth(2.0)
		ctx.SetStrokeStyle(gg.NewSolidPattern(color.RGBA{R: 255, G: 255, B: 0, A: 255}))
		ctx.Stroke()
	}
}
