package repository

import (
	"context"
	"time"
)

type CacheRepository interface {
	Get(ctx context.Context, cacheKey string) (string, error)
	Set(ctx context.Context, cacheKey string, cacheValue []byte, ttl time.Duration) error
	Delete(ctx context.Context, cacheKey string) error
	Close() error
}

type CacheFactory interface {
	Register(cacheType string, handler CacheRepository)
	GetHandler(cacheType string) (CacheRepository, bool)
}

type cacheFactory struct {
	registry map[string]CacheRepository
}

func NewCacheFactory() CacheFactory {
	return &cacheFactory{
		registry: make(map[string]CacheRepository),
	}
}

func (c *cacheFactory) Register(cacheType string, handler CacheRepository) {
	c.registry[cacheType] = handler
}

func (c *cacheFactory) GetHandler(cacheType string) (CacheRepository, bool) {
	handler, exists := c.registry[cacheType]
	return handler, exists
}
