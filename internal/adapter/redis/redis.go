package redis

import (
	"context"
	"time"

	"github.com/CSKU-Lab/cache/configs"
	"github.com/CSKU-Lab/cache/constants"
	"github.com/CSKU-Lab/cache/domain/repository"
	"github.com/redis/go-redis/v9"
)

type redisCacheAdapter struct {
	client *redis.Client
}

func NewRedisCacheAdapter(cfg *configs.Config) (repository.CacheRepository, error) {
	if cfg == nil {
		return nil, constants.CONFIG_NOT_FOUND
	}

	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		Protocol: 2,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &redisCacheAdapter{
		client: client,
	}, nil
}

func (r *redisCacheAdapter) Get(ctx context.Context, cacheKey string) (string, error) {
	val, err := r.client.Get(ctx, cacheKey).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

func (r *redisCacheAdapter) Set(ctx context.Context, cacheKey string, cacheValue []byte, ttl time.Duration) error {
	err := r.client.Set(ctx, cacheKey, cacheValue, ttl).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *redisCacheAdapter) Delete(ctx context.Context, cacheKey string) error {
	return r.client.Del(ctx, cacheKey).Err()
}

func (r *redisCacheAdapter) Close() error {
	return r.client.Close()
}
