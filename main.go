package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/CSKU-Lab/cache/constants"
	"github.com/CSKU-Lab/cache/domain/repository"
	"github.com/CSKU-Lab/cache/internal/adapter/redis"
	redisLib "github.com/redis/go-redis/v9"
)

type CacheApp interface {
	GetRepo() repository.CacheRepository
	Close() error
}

type cacheApp struct {
	repo repository.CacheRepository
}

type RedisOptions struct {
	Addr     string
	Password string
}

func NewRedis(rawOpts *RedisOptions) (CacheApp, error) {
	redisRepo, err := redis.NewRedisCacheAdapter(&redisLib.Options{
		Addr:     rawOpts.Addr,
		Password: rawOpts.Password,
		DB:       0,
		Protocol: 2,
	})
	if err != nil {
		return nil, err
	}
	return &cacheApp{
		repo: redisRepo,
	}, nil
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

type CacheBuild[T any] interface {
	All() CacheInstance[T]
	One(id string) CacheInstance[T]
	InvalidateAll(ctx context.Context) error
}

type cacheBuild[T any] struct {
	rawKey string
	ttl    time.Duration
	repo   repository.CacheRepository
}

type CacheInstance[T any] interface {
	GetFromCache(ctx context.Context) (*T, error)
	SetToCache(ctx context.Context, value T) error
	DeleteCache(ctx context.Context) error
	LazyCaching(ctx context.Context, fetch func() (T, error)) (T, error)
}
type cacheInstance[T any] struct {
	key    string
	rawKey string
	ttl    time.Duration
	repo   repository.CacheRepository
}

func NewCacheBuild[T any](rawKey string, ttl time.Duration, repo repository.CacheRepository) CacheBuild[T] {
	return &cacheBuild[T]{
		rawKey: rawKey,
		ttl:    ttl,
		repo:   repo,
	}
}

func (cb *cacheBuild[T]) One(id string) CacheInstance[T] {
	return &cacheInstance[T]{
		rawKey: cb.rawKey,
		key:    cb.rawKey + ":id:" + id,
		ttl:    cb.ttl,
		repo:   cb.repo,
	}
}

func (cb *cacheBuild[T]) All() CacheInstance[T] {
	return &cacheInstance[T]{
		rawKey: cb.rawKey,
		key:    cb.rawKey + ":all",
		ttl:    cb.ttl,
		repo:   cb.repo,
	}
}

func (cb *cacheBuild[T]) InvalidateAll(ctx context.Context) error {
	keys, err := cb.repo.SMembers(ctx, cb.rawKey+":index")
	if err != nil {
		return err
	}

	for _, k := range keys {
		err = cb.repo.Delete(ctx, string(k))
		if err != nil {
			return err
		}
	}

	err = cb.repo.Delete(ctx, cb.rawKey+":index")
	if err != nil {
		return err
	}

	return nil
}

func (ci *cacheInstance[T]) GetFromCache(ctx context.Context) (*T, error) {
	data, hit, err := ci.repo.Get(ctx, ci.key)
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

	err = ci.repo.Set(ctx, ci.key, jsonString, ci.ttl)
	if err != nil {
		return err
	}

	return ci.repo.SAdd(ctx, ci.rawKey+":index", []byte(ci.key))
}

func (ci *cacheInstance[T]) DeleteCache(ctx context.Context) error {
	err := ci.repo.Delete(ctx, ci.key)
	if err != nil {
		return err
	}
	return nil
}

func (ci *cacheInstance[T]) LazyCaching(
	ctx context.Context,
	fetch func() (T, error),
) (T, error) {
	var zero T

	res, err := ci.GetFromCache(ctx)
	if err != nil {
		return zero, err
	}

	if res != nil {
		return *res, nil
	}

	fetchedData, err := fetch()
	if err != nil {
		return zero, err
	}

	if err := ci.SetToCache(ctx, fetchedData); err != nil {
		return zero, err
	}

	return fetchedData, nil
}
