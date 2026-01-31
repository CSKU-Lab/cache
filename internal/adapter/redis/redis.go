package redis

import (
	"context"
	"time"

	"github.com/CSKU-Lab/cache/domain/repository"
	"github.com/redis/go-redis/v9"
)

type redisCacheAdapter struct {
	client *redis.Client
}

func NewRedisCacheAdapter(opts *redis.Options) (repository.CacheRepository, error) {
	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &redisCacheAdapter{
		client: client,
	}, nil
}

func (r *redisCacheAdapter) Get(ctx context.Context, cacheKey string) ([]byte, bool, error) {
	val, err := r.client.Get(ctx, cacheKey).Bytes()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	return val, true, nil
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
