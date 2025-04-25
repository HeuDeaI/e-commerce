package category

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"e-commerce/internal/domains"
	"github.com/redis/go-redis/v9"
)

type CachedCategoryRepository interface {
	GetCategoryByID(ctx context.Context, id uint) (*domains.Category, error)
	GetAllCategories(ctx context.Context) ([]*domains.Category, error)
	SetCategory(ctx context.Context, category *domains.Category, ttl time.Duration) error
	DeleteCache(ctx context.Context, key string) error
	SetAllCategories(ctx context.Context, categories []*domains.Category, ttl time.Duration) error
}

type cachedCategoryRepository struct {
	cache *redis.Client
}

func NewCachedCategoryRepository(cache *redis.Client) CachedCategoryRepository {
	return &cachedCategoryRepository{cache: cache}
}

func (r *cachedCategoryRepository) GetCategoryByID(ctx context.Context, id uint) (*domains.Category, error) {
	cacheKey := fmt.Sprintf("category:%d", id)
	var category domains.Category

	data, err := r.cache.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
		return nil, err
	} else if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(data), &category); err != nil {
		return nil, err
	}

	return &category, nil
}

func (r *cachedCategoryRepository) GetAllCategories(ctx context.Context) ([]*domains.Category, error) {
	cacheKey := "categories:all"
	var categories []*domains.Category

	data, err := r.cache.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
		return nil, err
	} else if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(data), &categories); err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *cachedCategoryRepository) SetCategory(ctx context.Context, category *domains.Category, ttl time.Duration) error {
	cacheKey := fmt.Sprintf("category:%d", category.ID)
	data, err := json.Marshal(category)
	if err != nil {
		return err
	}

	return r.cache.Set(ctx, cacheKey, data, ttl).Err()
}

func (r *cachedCategoryRepository) DeleteCache(ctx context.Context, key string) error {
	return r.cache.Del(ctx, key).Err()
}

func (r *cachedCategoryRepository) SetAllCategories(ctx context.Context, categories []*domains.Category, ttl time.Duration) error {
	cacheKey := "categories:all"
	data, err := json.Marshal(categories)
	if err != nil {
		return err
	}

	return r.cache.Set(ctx, cacheKey, data, ttl).Err()
}
