package redis

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

const QUERY_TIMEOUT = 5 * time.Second

func connectRedis() *redis.Client {

	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_URL") + ":" + os.Getenv("REDIS_PORT"),
	})

	return rdb
}

func PushListToRedis(pictureLinks []string, keyName string) error {
	rdb := connectRedis()
	defer rdb.Close()
	ctx, cancel := context.WithTimeout(context.Background(), QUERY_TIMEOUT)
	defer cancel()
	if err := rdb.RPush(ctx, keyName, pictureLinks).Err(); err != nil {
		return fmt.Errorf("Failed to push list to Redis: %w\n", err)
	}

	return nil
}

func ReadListFromRedis(keyName string) ([]string, error) {
	rdb := connectRedis()
	defer rdb.Close()
	ctx, cancel := context.WithTimeout(context.Background(), QUERY_TIMEOUT)
	defer cancel()
	vals, err := rdb.LRange(ctx, keyName, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("Failed to read list from Redis: %w\n", err)
	}
	return vals, nil
}
