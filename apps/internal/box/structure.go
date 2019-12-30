package box

// Box is an element within a picture
type Box struct {
	Element        string
	Confidence     float64
	X0, Y0, X1, Y1 int
}
