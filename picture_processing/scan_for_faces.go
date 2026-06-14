package picture_processing

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"photo_fetch/config"
	"photo_fetch/database"
	"strconv"
	"strings"

	"github.com/pgvector/pgvector-go"
)

type BoundingBox struct {
	X1 float64 `json:"x1"`
	Y1 float64 `json:"y1"`
	X2 float64 `json:"x2"`
	Y2 float64 `json:"y2"`
}

type FaceDetection struct {
	BoundingBox BoundingBox  `json:"boundingBox"`
	Embedding   RawEmbedding `json:"embedding"`
	Score       float64      `json:"score"`
}

// RawEmbedding handles the embedding being returned as either a JSON string or array
type RawEmbedding []float32

func (e *RawEmbedding) UnmarshalJSON(data []byte) error {
	// If it's a quoted string, strip the quotes and parse the inner array
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
	// Otherwise treat it as a normal JSON array
	var arr []float32
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}
	*e = arr
	return nil
}

type PredictResponse struct {
	FacialRecognition []FaceDetection `json:"facial-recognition"`
	ImageHeight       int             `json:"imageHeight"`
	ImageWidth        int             `json:"imageWidth"`
}

func detectFaces(imagePath string, minScore float64) (*PredictResponse, error) {
	entries := map[string]any{
		"facial-recognition": map[string]any{
			"detection": map[string]any{
				"modelName": config.ML_FACE_MODEL,
				"options": map[string]any{
					"minScore": minScore,
				},
			},
			"recognition": map[string]any{
				"modelName": config.ML_FACE_MODEL,
				"options":   map[string]any{},
			},
		},
	}

	entriesJSON, err := json.Marshal(entries)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal entries: %w", err)
	}

	imageFile, err := os.Open(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open image: %w", err)
	}
	defer imageFile.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	if err := writer.WriteField("entries", string(entriesJSON)); err != nil {
		return nil, fmt.Errorf("failed to write entries field: %w", err)
	}

	part, err := writer.CreateFormFile("image", filepath.Base(imagePath))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := io.Copy(part, imageFile); err != nil {
		return nil, fmt.Errorf("failed to copy image data: %w", err)
	}

	writer.Close()

	url := fmt.Sprintf("http://%s:3003/predict", config.ML_SERVER)
	req, err := http.NewRequest(http.MethodPost, url, &body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, body)
	}

	var result PredictResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func getPersonNameByEmbedding(ctx context.Context, e RawEmbedding) (string, error) {

	db, _ := database.ConnectToDB()

	// 1. Use $1 as a placeholder instead of string interpolation
	query := `
		SELECT p."name"
		FROM (
			SELECT f."faceId"
			FROM face_search f
			ORDER BY f.embedding <=> $1
			LIMIT 1
		) cf
		LEFT JOIN public.asset_face af ON cf."faceId" = af.id  
		LEFT JOIN public.person p ON p.id = af."personId";
	`

	var name string
	err := db.QueryRowContext(ctx, query, pgvector.NewVector(e)).Scan(&name)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil // No match found
		}
		return "", err // Database error
	}

	fmt.Printf("Found person %s\n", name)

	return name, nil
}

func ProcessImages() {
	fmt.Println("Processing FACE")
	result, err := detectFaces(`C:\Users\jakeb\OneDrive\Pictures\Camera Roll\WIN_20260311_11_36_37_Pro.jpg`, 0.7)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Image size: %dx%d\n", result.ImageWidth, result.ImageHeight)
	fmt.Printf("Faces detected: %d\n", len(result.FacialRecognition))
	if len(result.FacialRecognition) == 0 {
		return
	}
	for i, face := range result.FacialRecognition {
		fmt.Printf("  Face %d: score=%.2f bbox=(%.0f,%.0f)-(%.0f,%.0f) embedding_dims=%d\n",
			i+1, face.Score,
			face.BoundingBox.X1, face.BoundingBox.Y1,
			face.BoundingBox.X2, face.BoundingBox.Y2,
			len(face.Embedding),
		)
		getPersonNameByEmbedding(context.Background(), face.Embedding)
	}
}
