package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/CSKU-Lab/cache/configs"
	"github.com/CSKU-Lab/cache/constants"
	"github.com/CSKU-Lab/cache/domain/repository"
	"github.com/CSKU-Lab/cache/internal/adapter/redis"
	redisLib "github.com/redis/go-redis/v9"
)

type CacheApp interface {
	NewRedis(opts *redisLib.Options) (repository.CacheRepository, error)
	Close() error
}

type cacheApp struct {
	cfg  *configs.Config
	repo repository.CacheRepository
}

type RedisOptions struct {
	Addr     string
	Password string
}

func (ca *cacheApp) NewRedis(rawOpts *RedisOptions) (repository.CacheRepository, error) {
	redisRepo, err := redis.NewRedisCacheAdapter(&redisLib.Options{
		Addr:     rawOpts.Addr,
		Password: rawOpts.Password,
		DB:       0,
		Protocol: 2,
	})
	if err != nil {
		return nil, err
	}
	ca.repo = redisRepo
	return redisRepo, nil
}

func (ca *cacheApp) Close() error {
	if ca.repo == nil {
		return constants.NO_CACHE_CONN
	}

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
