package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var IMMICH_API_KEY string
var IMMICH_RO_API_KEY string
var IMMICH_URL string
var DB_HOST string
var DB_PORT string
var DB_USER string
var DB_PASSWORD string
var DB_NAME string
var REDIS_URL string
var REDIS_PORT string
var PHOTOS_LIST_KEY string
var GO_LISTEN_PORT string
var SLIDESHOW_BATCH_SIZE int
var DOORBELL_HOST string
var UNIFI_API_KEY string
var DOORBELL_ID string
var ML_SERVER string
var ML_FACE_MODEL string

func LoadConfig() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	IMMICH_API_KEY = os.Getenv("IMMICH_API_KEY")
	IMMICH_RO_API_KEY = os.Getenv("IMMICH_RO_API_KEY")
	IMMICH_URL = os.Getenv("IMMICH_URL")
	DB_HOST = os.Getenv("DB_HOST")
	DB_PORT = os.Getenv("DB_PORT")
	DB_USER = os.Getenv("DB_USER")
	DB_PASSWORD = os.Getenv("DB_PASSWORD")
	DB_NAME = os.Getenv("DB_NAME")
	REDIS_URL = os.Getenv("REDIS_URL")
	REDIS_PORT = os.Getenv("REDIS_PORT")
	PHOTOS_LIST_KEY = os.Getenv("PHOTOS_LIST_KEY")
	GO_LISTEN_PORT = os.Getenv("GO_LISTEN_PORT")
	SLIDESHOW_BATCH_SIZE, _ = strconv.Atoi(os.Getenv("SLIDESHOW_BATCH_SIZE"))
	DOORBELL_HOST = os.Getenv("DOORBELL_HOST")
	UNIFI_API_KEY = os.Getenv("UNIFI_API_KEY")
	DOORBELL_ID = os.Getenv("DOORBELL_ID")
	ML_SERVER = os.Getenv("ML_SERVER")
	ML_FACE_MODEL = os.Getenv("ML_FACE_MODEL")

}
