package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheRepository[T any] interface {
	GetByID(ctx context.Context, id int) (*T, error)
	GetByKey(ctx context.Context, key string) ([]*T, error)
	GetAll(ctx context.Context) ([]*T, error)
	SetByID(ctx context.Context, id int, item *T) error
	SetByKey(ctx context.Context, key string, items []*T) error
	SetAll(ctx context.Context, items []*T) error
	Delete(ctx context.Context, id int) error
	DeleteAll(ctx context.Context) error
}

type cacheRepository[T any] struct {
	client    *redis.Client
	keyPrefix string
	ttl       time.Duration
}

func NewCacheRepository[T any](client *redis.Client, keyPrefix string) CacheRepository[T] {
	return &cacheRepository[T]{
		client:    client,
		keyPrefix: keyPrefix,
		ttl:       20 * time.Minute,
	}
}

func (r *cacheRepository[T]) GetByID(ctx context.Context, id int) (*T, error) {
	key := fmt.Sprintf("%s:%d", r.keyPrefix, id)
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var item T
	if err := json.Unmarshal([]byte(data), &item); err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *cacheRepository[T]) GetByKey(ctx context.Context, key string) ([]*T, error) {
	cacheKey := fmt.Sprintf("%s:%s", r.keyPrefix, key)
	data, err := r.client.Get(ctx, cacheKey).Result()
	if err != nil {
		return nil, err
	}

	var items []*T
	if err := json.Unmarshal([]byte(data), &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *cacheRepository[T]) GetAll(ctx context.Context) ([]*T, error) {
	data, err := r.client.Get(ctx, r.keyPrefix+":all").Result()
	if err != nil {
		return nil, err
	}

	var items []*T
	if err := json.Unmarshal([]byte(data), &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *cacheRepository[T]) SetByID(ctx context.Context, id int, item *T) error {
	key := fmt.Sprintf("%s:%d", r.keyPrefix, id)
	data, err := json.Marshal(item)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, r.ttl).Err()
}

func (r *cacheRepository[T]) SetByKey(ctx context.Context, key string, items []*T) error {
	cacheKey := fmt.Sprintf("%s:%s", r.keyPrefix, key)
	data, err := json.Marshal(items)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, cacheKey, data, r.ttl).Err()
}

func (r *cacheRepository[T]) SetAll(ctx context.Context, items []*T) error {
	data, err := json.Marshal(items)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.keyPrefix+":all", data, r.ttl).Err()
}

func (r *cacheRepository[T]) Delete(ctx context.Context, id int) error {
	key := fmt.Sprintf("%s:%d", r.keyPrefix, id)
	return r.client.Del(ctx, key).Err()
}

func (r *cacheRepository[T]) DeleteAll(ctx context.Context) error {
	return r.client.Del(ctx, r.keyPrefix+":all").Err()
}
