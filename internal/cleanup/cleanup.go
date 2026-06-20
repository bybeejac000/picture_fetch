package cleanup

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func StartCleanupLoop(ctx context.Context, dir string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			safeCleanFolder(dir)
		}
	}
}

func safeCleanFolder(dir string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("cleanup: recovered from panic: %v", r)
		}
	}()
	cleanFolder(dir)
}

func cleanFolder(dir string) {
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("cleanup: failed to read directory: %v", err)
		return
	}

	for _, file := range files {
		filePath := filepath.Join(dir, file.Name())
		if file.IsDir() {
			cleanFolder(filePath)
		} else {
			timestamp, err := parseTimestamp(file.Name())
			if err != nil {
				os.Remove(filePath)
				log.Printf("cleanup: failed to parse timestamp: %v", err)
				continue
			}
			if timestamp.Before(time.Now().Add(-10 * time.Minute)) {
				log.Printf("cleanup: cleaning up old file: %s, timestamp difference: %v", filePath, time.Since(timestamp))
				err := os.Remove(filePath)
				if err != nil {
					log.Printf("cleanup: failed to remove file: %v", err)
				}
			}
		}
	}
}

func parseTimestamp(filename string) (time.Time, error) {
	name := strings.TrimSuffix(filename, ".jpg")

	parts := strings.Split(name, "_")
	if len(parts) < 2 {
		return time.Time{}, fmt.Errorf("unexpected filename format: %s", filename)
	}

	dateTimeStr := parts[len(parts)-2] + parts[len(parts)-1]

	loc, err := time.LoadLocation("America/Denver")
	if err != nil {
		return time.Time{}, err
	}

	return time.ParseInLocation("20060102150405", dateTimeStr, loc)
}
