package startup

import (
	"context"
	"log"
	"photo_fetch/config"
	"photo_fetch/redis"
	"photo_fetch/routes"
	"time"
)

func StartupScript() {
	// Load configuration
	config.LoadConfig()

	// Clear the Redis list at startup
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redis.ClearListInRedis(ctx, config.PHOTOS_LIST_KEY); err != nil {
		log.Fatal(err)
	}

	// Populate the redis list randomly
	err := routes.RefreshSlideshowList(ctx)
	if err != nil {
		log.Fatal(err)
	}

}
