package faces

import (
	"encoding/json"
	"strconv"
	"strings"
)

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

// RawEmbedding handles the embedding being returned as either a JSON string or array.
type RawEmbedding []float32

func (e *RawEmbedding) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		s = strings.TrimSpace(s)
		s = strings.TrimPrefix(s, "[")
		s = strings.TrimSuffix(s, "]")
		for _, part := range strings.Split(s, ",") {
			f, err := strconv.ParseFloat(strings.TrimSpace(part), 32)
			if err != nil {
				return err
			}
			*e = append(*e, float32(f))
		}
		return nil
	}

	var arr []float32
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}
	*e = arr
	return nil
}
