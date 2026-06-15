package store

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"photo_fetch/internal/config"
)

const queryTimeout = 5 * time.Second

type Store struct {
	rdb *redis.Client
}

func New(cfg *config.Config) *Store {
	return &Store{
		rdb: redis.NewClient(&redis.Options{
			Addr: cfg.RedisURL + ":" + cfg.RedisPort,
		}),
	}
}

func (s *Store) Close() error {
	return s.rdb.Close()
}

func (s *Store) PushList(ctx context.Context, links []string, key string) error {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	if err := s.rdb.RPush(ctx, key, links).Err(); err != nil {
		return fmt.Errorf("pushing list to redis: %w", err)
	}
	return nil
}

func (s *Store) ReadList(ctx context.Context, key string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	vals, err := s.rdb.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("reading list from redis: %w", err)
	}
	return vals, nil
}

func (s *Store) ClearList(ctx context.Context, key string) error {
	_, err := s.rdb.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Del(ctx, key)
		return nil
	})
	if err != nil {
		return fmt.Errorf("clearing list in redis: %w", err)
	}
	return nil
}
