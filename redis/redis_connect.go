package redis

import (
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
