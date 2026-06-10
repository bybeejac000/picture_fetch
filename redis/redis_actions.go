package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func PushListToRedis(ctx context.Context, pictureLinks []string, keyName string) error {
	rdb := connectRedis()
	defer rdb.Close()
	ctx, cancel := context.WithTimeout(ctx, QUERY_TIMEOUT)
	defer cancel()
	if err := rdb.RPush(ctx, keyName, pictureLinks).Err(); err != nil {
		return fmt.Errorf("Failed to push list to Redis: %w\n", err)
	}

	return nil
}

func ReadListFromRedis(ctx context.Context, keyName string) ([]string, error) {
	rdb := connectRedis()
	defer rdb.Close()
	ctx, cancel := context.WithTimeout(ctx, QUERY_TIMEOUT)
	defer cancel()
	vals, err := rdb.LRange(ctx, keyName, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("Failed to read list from Redis: %w\n", err)
	}
	return vals, nil
}

func ClearListInRedis(ctx context.Context, keyName string) error {
	rdb := connectRedis()
	defer rdb.Close()
	ctx, cancel := context.WithTimeout(ctx, QUERY_TIMEOUT)
	defer cancel()
	_, err := rdb.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Del(ctx, keyName)
		return nil
	})
	if err != nil {
		return fmt.Errorf("Failed to clear list in Redis: %w\n", err)
	}
	return nil
}
