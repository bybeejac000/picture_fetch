package main

import (
	"fmt"
	"log"
	"photo_fetch/database"
)

func main() {
	envConfig := declareEnv()

	links, err := database.ReturnSlideshowList(envConfig.ImmichURL, envConfig.ImmichAPIKey)
	if err != nil {
		log.Fatalf("Failed to get slideshow list: %v", err)
	}

	fmt.Println(links)
}
