package main

import (
	"log"
	"photo_fetch/database"
	"photo_fetch/redis"
)

func main() {
	envConfig := declareEnv()

	links, err := database.ReturnSlideshowList(envConfig.ImmichURL, envConfig.ImmichAPIKey)
	if err != nil {
		log.Fatalf("Failed to get slideshow list: %v", err)
	}

	err = redis.PushListToRedis(links, envConfig.PhotosListKey)
	if err != nil {
		log.Fatalf("Failed to push list to Redis: %v", err)
	}

}
