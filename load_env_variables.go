package main

import (
	"os"
	"strings"

	_ "github.com/joho/godotenv/autoload"
)

type EnvConfig struct {
	ImmichURL    string
	ImmichAPIKey string
}

func declareEnv() EnvConfig {
	immichURL := strings.TrimSuffix(os.Getenv("IMMICH_URL"), "/")
	immichApiKey := os.Getenv("IMMICH_API_KEY")
	return EnvConfig{
		ImmichURL:    immichURL,
		ImmichAPIKey: immichApiKey,
	}
}
