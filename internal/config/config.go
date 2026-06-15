package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ImmichAPIKey       string
	ImmichROAPIKey     string
	ImmichURL          string
	DBHost             string
	DBPort             string
	DBUser             string
	DBPassword         string
	DBName             string
	RedisURL           string
	RedisPort          string
	PhotosListKey      string
	GoListenPort       string
	SlideshowBatchSize int
	DoorbellHost       string
	UnifiAPIKey        string
	DoorbellID         string
	MLServer           string
	MLFaceModel        string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("loading .env file: %w", err)
	}

	batchSize, _ := strconv.Atoi(os.Getenv("SLIDESHOW_BATCH_SIZE"))

	return &Config{
		ImmichAPIKey:       os.Getenv("IMMICH_API_KEY"),
		ImmichROAPIKey:     os.Getenv("IMMICH_RO_API_KEY"),
		ImmichURL:          os.Getenv("IMMICH_URL"),
		DBHost:             os.Getenv("DB_HOST"),
		DBPort:             os.Getenv("DB_PORT"),
		DBUser:             os.Getenv("DB_USER"),
		DBPassword:         os.Getenv("DB_PASSWORD"),
		DBName:             os.Getenv("DB_NAME"),
		RedisURL:           os.Getenv("REDIS_URL"),
		RedisPort:          os.Getenv("REDIS_PORT"),
		PhotosListKey:      os.Getenv("PHOTOS_LIST_KEY"),
		GoListenPort:       os.Getenv("GO_LISTEN_PORT"),
		SlideshowBatchSize: batchSize,
		DoorbellHost:       os.Getenv("DOORBELL_HOST"),
		UnifiAPIKey:        os.Getenv("UNIFI_API_KEY"),
		DoorbellID:         os.Getenv("DOORBELL_ID"),
		MLServer:           os.Getenv("ML_SERVER"),
		MLFaceModel:        os.Getenv("ML_FACE_MODEL"),
	}, nil
}
