package repository

import (
	"context"
	"time"
)

type CacheRepository interface {
	Get(ctx context.Context, cacheKey string) ([]byte, bool, error)
	Set(ctx context.Context, cacheKey string, cacheValue []byte, ttl time.Duration) error
	Delete(ctx context.Context, cacheKey string) error
	Close() error

	SAdd(ctx context.Context, key string, member []byte) error
	SMembers(ctx context.Context, key string) ([][]byte, error)
}
