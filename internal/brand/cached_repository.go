package brand

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"e-commerce/internal/domains"
	"github.com/redis/go-redis/v9"
)

type CachedBrandRepository interface {
	GetBrandByID(ctx context.Context, id uint) (*domains.Brand, error)
	GetAllBrands(ctx context.Context) ([]*domains.Brand, error)
	SetBrand(ctx context.Context, brand *domains.Brand, ttl time.Duration) error
	DeleteCache(ctx context.Context, key string) error
	SetAllBrands(ctx context.Context, brands []*domains.Brand, ttl time.Duration) error
}

type cachedBrandRepository struct {
	cache *redis.Client
}

func NewCachedBrandRepository(cache *redis.Client) CachedBrandRepository {
	return &cachedBrandRepository{cache: cache}
}

func (r *cachedBrandRepository) GetBrandByID(ctx context.Context, id uint) (*domains.Brand, error) {
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

func (r *cachedBrandRepository) GetAllBrands(ctx context.Context) ([]*domains.Brand, error) {
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

func (r *cachedBrandRepository) SetBrand(ctx context.Context, brand *domains.Brand, ttl time.Duration) error {
	cacheKey := fmt.Sprintf("brand:%d", brand.ID)
	data, err := json.Marshal(brand)
	if err != nil {
		return err
	}

	return r.cache.Set(ctx, cacheKey, data, ttl).Err()
}

func (r *cachedBrandRepository) DeleteCache(ctx context.Context, key string) error {
	return r.cache.Del(ctx, key).Err()
}

func (r *cachedBrandRepository) SetAllBrands(ctx context.Context, brands []*domains.Brand, ttl time.Duration) error {
	cacheKey := "brands:all"
	data, err := json.Marshal(brands)
	if err != nil {
		return err
	}

	return r.cache.Set(ctx, cacheKey, data, ttl).Err()
}
