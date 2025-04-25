package product

import (
	"context"
	"e-commerce/internal/domains"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type CachedProductRepository interface {
	GetProductByID(ctx context.Context, id uint) (*domains.Product, error)
	GetAllProducts(ctx context.Context) ([]*domains.Product, error)
	SetProduct(ctx context.Context, product *domains.Product, ttl time.Duration) error
	DeleteCache(ctx context.Context, key string) error
	SetAllProducts(ctx context.Context, products []*domains.Product, ttl time.Duration) error
}

type cachedProductRepository struct {
	cache *redis.Client
}

func NewCachedProductRepository(cache *redis.Client) CachedProductRepository {
	return &cachedProductRepository{cache: cache}
}

func (r *cachedProductRepository) GetProductByID(ctx context.Context, id uint) (*domains.Product, error) {
	cacheKey := fmt.Sprintf("product:%d", id)
	var product domains.Product

	data, err := r.cache.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
		return nil, err
	} else if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(data), &product); err != nil {
		return nil, err
	}

	return &product, nil
}

func (r *cachedProductRepository) GetAllProducts(ctx context.Context) ([]*domains.Product, error) {
	cacheKey := "products:all"
	var products []*domains.Product

	data, err := r.cache.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
		return nil, err
	} else if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(data), &products); err != nil {
		return nil, err
	}

	return products, nil
}

func (r *cachedProductRepository) SetProduct(ctx context.Context, product *domains.Product, ttl time.Duration) error {
	cacheKey := fmt.Sprintf("product:%d", product.ID)
	data, err := json.Marshal(product)
	if err != nil {
		return err
	}

	return r.cache.Set(ctx, cacheKey, data, ttl).Err()
}

func (r *cachedProductRepository) DeleteCache(ctx context.Context, key string) error {
	return r.cache.Del(ctx, key).Err()
}

func (r *cachedProductRepository) SetAllProducts(ctx context.Context, products []*domains.Product, ttl time.Duration) error {
	cacheKey := "products:all"
	data, err := json.Marshal(products)
	if err != nil {
		return err
	}

	return r.cache.Set(ctx, cacheKey, data, ttl).Err()
}
