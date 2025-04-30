package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"e-commerce/internal/domains"
	"github.com/redis/go-redis/v9"
)

type BaseCachedRepository interface {
	GetByID(ctx context.Context, id uint) (*domains.Brand, error)
	GetAll(ctx context.Context) ([]*domains.Brand, error)
	Set(ctx context.Context, brand *domains.Brand, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	SetAll(ctx context.Context, brands []*domains.Brand, ttl time.Duration) error
}

type baseCachedRepository struct {
	cache *redis.Client
}

func NewBaseCachedRepository(cache *redis.Client) BaseCachedRepository {
	return &baseCachedRepository{cache: cache}
}

func (r *baseCachedRepository) GetByID(ctx context.Context, id uint) (*domains.Brand, error) {
	cacheKey := fmt.Sprintf("brand:%d", id)
	var brand domains.Brand

	data, err := r.cache.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
		return nil, err
	} else if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(data), &brand); err != nil {
		return nil, err
	}

	return &brand, nil
}

func (r *baseCachedRepository) GetAll(ctx context.Context) ([]*domains.Brand, error) {
	cacheKey := "brands:all"
	var brands []*domains.Brand

	data, err := r.cache.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
		return nil, err
	} else if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(data), &brands); err != nil {
		return nil, err
	}

	return brands, nil
}

func (r *baseCachedRepository) Set(ctx context.Context, brand *domains.Brand, ttl time.Duration) error {
	cacheKey := fmt.Sprintf("brand:%d", brand.ID)
	data, err := json.Marshal(brand)
	if err != nil {
		return err
	}

	return r.cache.Set(ctx, cacheKey, data, ttl).Err()
}

func (r *baseCachedRepository) Delete(ctx context.Context, key string) error {
	return r.cache.Del(ctx, key).Err()
}

func (r *baseCachedRepository) SetAll(ctx context.Context, brands []*domains.Brand, ttl time.Duration) error {
	cacheKey := "brands:all"
	data, err := json.Marshal(brands)
	if err != nil {
		return err
	}

	return r.cache.Set(ctx, cacheKey, data, ttl).Err()
}
