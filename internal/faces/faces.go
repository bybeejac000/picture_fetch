package faces

import (
	"bytes"
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/fogleman/gg"
	"github.com/pgvector/pgvector-go"

	"photo_fetch/internal/config"
)

const outputDir = "doorbell_frames"

type Processor struct {
	db  *sql.DB
	cfg *config.Config
}

func NewProcessor(db *sql.DB, cfg *config.Config) *Processor {
	return &Processor{db: db, cfg: cfg}
}

func (p *Processor) captureSnapshot(client *http.Client, label string) (string, error) {
	url := fmt.Sprintf("https://%s/proxy/protect/integration/v1/cameras/%s/snapshot", p.cfg.DoorbellHost, p.cfg.DoorbellID)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("X-API-Key", p.cfg.UnifiAPIKey)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 300))
		return "", fmt.Errorf("bad status %d: %s", resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response body: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := filepath.Join(outputDir, fmt.Sprintf("doorbell_%s_%s.jpg", label, timestamp))

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return "", fmt.Errorf("writing file: %w", err)
	}

	log.Printf("saved: %s (%.1f KB)", filename, float64(len(data))/1024)
	return filename, nil
}

func (p *Processor) Capture() (string, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("creating output directory: %w", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	log.Println("capturing snapshot")
	filePath, err := p.captureSnapshot(client, "frame1")
	if err != nil {
		return "", fmt.Errorf("capturing snapshot: %w", err)
	}

	log.Printf("done, frame saved to ./%s/", outputDir)
	return filePath, nil
}

func (p *Processor) detectFaces(imagePath string, minScore float64) (*FaceDetectionResponse, error) {
	entries := map[string]any{
		"facial-recognition": map[string]any{
			"detection": map[string]any{
				"modelName": p.cfg.MLFaceModel,
				"options": map[string]any{
					"minScore": minScore,
				},
			},
			"recognition": map[string]any{
				"modelName": p.cfg.MLFaceModel,
				"options":   map[string]any{},
			},
		},
	}

	entriesJSON, err := json.Marshal(entries)
	if err != nil {
		return nil, fmt.Errorf("marshaling entries: %w", err)
	}

	imageFile, err := os.Open(imagePath)
	if err != nil {
		return nil, fmt.Errorf("opening image: %w", err)
	}
	defer imageFile.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	if err := writer.WriteField("entries", string(entriesJSON)); err != nil {
		return nil, fmt.Errorf("writing entries field: %w", err)
	}

	part, err := writer.CreateFormFile("image", filepath.Base(imagePath))
	if err != nil {
		return nil, fmt.Errorf("creating form file: %w", err)
	}
	if _, err := io.Copy(part, imageFile); err != nil {
		return nil, fmt.Errorf("copying image data: %w", err)
	}

	writer.Close()

	url := fmt.Sprintf("http://%s:3003/predict", p.cfg.MLServer)
	req, err := http.NewRequest(http.MethodPost, url, &body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
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

	var result FaceDetectionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result, nil
}

func (p *Processor) personNameByEmbedding(ctx context.Context, e RawEmbedding) (string, error) {
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
	err := p.db.QueryRowContext(ctx, query, pgvector.NewVector(e)).Scan(&name)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}

	log.Printf("found person %s", name)
	return name, nil
}

func (p *Processor) addNames(filePath string, detections []FaceDetection, names []string, hasUnknown bool) error {
	img, err := gg.LoadImage(filePath)
	if err != nil {
		return fmt.Errorf("loading image: %w", err)
	}

	dc := gg.NewContextForImage(img)

	if hasUnknown {
		dc.SetRGB(1, 0, 0)
		dc.DrawRectangle(0, 0, float64(dc.Width()), 40)
		dc.Fill()
		dc.SetRGB(1, 1, 1)
		dc.DrawStringAnchored("New Face Detected!", float64(dc.Width())/2, 20, 0.5, 0.5)
	}

	for _, face := range detections {
		dc.SetRGB(0, 1, 0)
		dc.SetLineWidth(3)
		x := face.BoundingBox.X1
		y := face.BoundingBox.Y1
		w := face.BoundingBox.X2 - face.BoundingBox.X1
		h := face.BoundingBox.Y2 - face.BoundingBox.Y1
		dc.DrawRectangle(x, y, w, h)
		dc.Stroke()
	}

	dc.SetRGB(0, 1, 0)
	for i, name := range names {
		face := detections[i]
		x := face.BoundingBox.X1
		y := face.BoundingBox.Y1
		dc.DrawString(name, x, y-5)
	}

	return dc.SavePNG(filePath)
}

func (p *Processor) Process(filePath string) ([]string, bool) {
	log.Println("processing face")
	result, err := p.detectFaces(filePath, 0.5)
	if err != nil {
		log.Printf("error detecting faces: %v", err)
		log.Println("deleting file")
		os.Remove(filePath)
		return nil, false
	}

	log.Printf("image size: %dx%d", result.ImageWidth, result.ImageHeight)
	log.Printf("faces detected: %d", len(result.FacialRecognition))
	if len(result.FacialRecognition) == 0 {
		log.Println("no faces detected in this picture, deleting file")
		os.Remove(filePath)
		return nil, false
	}

	facesList := []string{}
	for i, face := range result.FacialRecognition {
		log.Printf("face %d: score=%.2f bbox=(%.0f,%.0f)-(%.0f,%.0f) embedding_dims=%d",
			i+1, face.Score,
			face.BoundingBox.X1, face.BoundingBox.Y1,
			face.BoundingBox.X2, face.BoundingBox.Y2,
			len(face.Embedding),
		)
		var e RawEmbedding
		if err := e.UnmarshalJSON([]byte(face.Embedding)); err != nil {
			log.Printf("failed to parse embedding: %v", err)
			continue
		}
		faceName, err := p.personNameByEmbedding(context.Background(), e)
		if err != nil {
			log.Printf("error: %v", err)
			continue
		}
		facesList = append(facesList, faceName)
	}
	p.addNames(filePath, result.FacialRecognition, facesList, true)

	return facesList, true
}
