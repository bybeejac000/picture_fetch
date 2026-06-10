package main

import (
	"context"
	"fmt"
	"os"
	"photo_fetch/database"
	"photo_fetch/redis"
)

func RefreshSlideshowList(ctx context.Context) error {
	links, err := database.ReturnSlideshowList(ctx, os.Getenv("IMMICH_URL"), os.Getenv("IMMICH_RO_API_KEY"))
	if err != nil {
		return fmt.Errorf("Failed to get slideshow list: %v", err)
	}
	err = redis.ClearListInRedis(ctx, os.Getenv("PHOTOS_LIST_KEY"))
	if err != nil {
		return fmt.Errorf("Failed to clear list in Redis: %v", err)
	}
	err = redis.PushListToRedis(ctx, links, os.Getenv("PHOTOS_LIST_KEY"))
	if err != nil {
		return fmt.Errorf("Failed to push list to Redis: %v", err)
	}
	return nil
}
