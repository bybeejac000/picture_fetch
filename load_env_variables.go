package main

import (
	"os"
	"strings"

	_ "github.com/joho/godotenv/autoload"
)

type EnvConfig struct {
	ImmichURL     string
	ImmichAPIKey  string
	PhotosListKey string
}

func declareEnv() EnvConfig {
	immichURL := strings.TrimSuffix(os.Getenv("IMMICH_URL"), "/")
	immichApiKey := os.Getenv("IMMICH_RO_API_KEY")
	photosListKey := os.Getenv("PHOTOS_LIST_KEY")
	return EnvConfig{
		ImmichURL:     immichURL,
		ImmichAPIKey:  immichApiKey,
		PhotosListKey: photosListKey,
	}
}
