package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type CachedRepositoryInterface[T any] interface {
	GetByID(ctx context.Context, id int) (*T, error)
	GetAll(ctx context.Context) ([]*T, error)
	Set(ctx context.Context, id int, item *T, ttl time.Duration) error
	SetAll(ctx context.Context, items []*T, ttl time.Duration) error
	Delete(ctx context.Context, id int) error
	DeleteAll(ctx context.Context) error
}

type BaseCachedRepository[T any] struct {
	client    *redis.Client
	keyPrefix string
}

func NewBaseCachedRepository[T any](client *redis.Client, keyPrefix string) *BaseCachedRepository[T] {
	return &BaseCachedRepository[T]{
		client:    client,
		keyPrefix: keyPrefix,
	}
}

func (r *BaseCachedRepository[T]) GetByID(ctx context.Context, id int) (*T, error) {
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

func (r *BaseCachedRepository[T]) GetAll(ctx context.Context) ([]*T, error) {
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

func (r *BaseCachedRepository[T]) Set(ctx context.Context, id int, item *T, ttl time.Duration) error {
	key := fmt.Sprintf("%s:%d", r.keyPrefix, id)
	data, err := json.Marshal(item)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *BaseCachedRepository[T]) SetAll(ctx context.Context, items []*T, ttl time.Duration) error {
	data, err := json.Marshal(items)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.keyPrefix+":all", data, ttl).Err()
}

func (r *BaseCachedRepository[T]) Delete(ctx context.Context, id int) error {
	key := fmt.Sprintf("%s:%d", r.keyPrefix, id)
	return r.client.Del(ctx, key).Err()
}

func (r *BaseCachedRepository[T]) DeleteAll(ctx context.Context) error {
	return r.client.Del(ctx, r.keyPrefix+":all").Err()
}
