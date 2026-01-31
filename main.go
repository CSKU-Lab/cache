package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/CSKU-Lab/cache/configs"
	"github.com/CSKU-Lab/cache/constants"
	"github.com/CSKU-Lab/cache/domain/repository"
	"github.com/CSKU-Lab/cache/internal/adapter/redis"
)

type CacheApp interface {
	GetRepo() repository.CacheRepository
	Close() error
}

type cacheApp struct {
	repo repository.CacheRepository
}

func Init(cacheVariant string) (CacheApp, error) {
	cfg := configs.NewConfig()
	redisRepo, err := redis.NewRedisCacheAdapter(cfg)
	if err != nil {
		return nil, err
	}

	cacheFactory := repository.NewCacheFactory()
	cacheFactory.Register("redis", redisRepo)

	cacheRepo, exists := cacheFactory.GetHandler(cacheVariant)
	if !exists {
		return nil, constants.CACHE_VARIANT_NOT_FOUND
	}

	return &cacheApp{
		repo: cacheRepo,
	}, nil
}

func (ca *cacheApp) Close() error {
	err := ca.repo.Close()
	if err != nil {
		return err
	}
	return nil
}

func (ca *cacheApp) GetRepo() repository.CacheRepository {
	return ca.repo
}

type CacheInstance[T any] interface {
	GetFromCache(ctx context.Context) (*T, error)
	SetToCache(ctx context.Context, value T) error
	DeleteCache(ctx context.Context) error
}
type cacheInstance[T any] struct {
	cache string
	ttl   time.Duration
	repo  repository.CacheRepository
}

func NewCacheInstance[T any](cache string, ttl time.Duration, repo repository.CacheRepository) CacheInstance[T] {
	return &cacheInstance[T]{
		cache: cache,
		ttl:   ttl,
		repo:  repo,
	}
}

func (ci *cacheInstance[T]) GetFromCache(ctx context.Context) (*T, error) {
	data, hit, err := ci.repo.Get(ctx, ci.cache)
	if err != nil {
		return nil, err
	}

	if !hit {
		return nil, nil
	}

	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (ci *cacheInstance[T]) SetToCache(ctx context.Context, value T) error {
	jsonString, err := json.Marshal(value)
	if err != nil {
		return err
	}

	err = ci.repo.Set(ctx, ci.cache, jsonString, ci.ttl)
	if err != nil {
		return err
	}

	return nil
}

func (ci *cacheInstance[T]) DeleteCache(ctx context.Context) error {
	err := ci.repo.Delete(ctx, ci.cache)
	if err != nil {
		return err
	}
	return nil
}
