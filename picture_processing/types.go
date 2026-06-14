package picture_processing

type BoundingBox struct {
	X1 float64 `json:"x1"`
	Y1 float64 `json:"y1"`
	X2 float64 `json:"x2"`
	Y2 float64 `json:"y2"`
}

type FaceDetection struct {
	BoundingBox BoundingBox `json:"boundingBox"`
	Embedding   string      `json:"embedding"`
	Score       float64     `json:"score"`
}

type FaceDetectionResponse struct {
	FacialRecognition []FaceDetection `json:"facial-recognition"`
	ImageHeight       int             `json:"imageHeight"`
	ImageWidth        int             `json:"imageWidth"`
}
