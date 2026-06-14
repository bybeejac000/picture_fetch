package picture_processing

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"photo_fetch/config"
	"time"
)

const (
	outputDir = "doorbell_frames"
)

func captureSnapshot(client *http.Client, label string) (string, error) {
	url := fmt.Sprintf("https://%s/proxy/protect/integration/v1/cameras/%s/snapshot", config.DOORBELL_HOST, config.DOORBELL_ID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("X-API-Key", config.UNIFI_API_KEY)

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
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := filepath.Join(outputDir, fmt.Sprintf("doorbell_%s_%s.jpg", label, timestamp))

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("  Saved: %s (%.1f KB)\n", filename, float64(len(data))/1024)
	return filename, nil
}

func CaptureFrames() (string, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create output directory: %v\n", err)
		os.Exit(1)
	}

	// Skip TLS verification (equivalent to verify=False in Python)
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	fmt.Println("Capturing snapshot...")
	filePath, err := captureSnapshot(client, "frame1")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nDone. Frame saved to ./%s/\n", outputDir)
	return filePath, nil
}
