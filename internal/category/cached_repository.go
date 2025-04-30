// internal/category/cache.go
package category

import (
	"context"
	"time"

	"e-commerce/internal/cache"
	"e-commerce/internal/domains"
	"github.com/redis/go-redis/v9"
)

type CachedCategoryRepository interface {
	GetCategoryByID(ctx context.Context, id int) (*domains.Category, error)
	GetAllCategories(ctx context.Context) ([]*domains.Category, error)
	SetCategory(ctx context.Context, category *domains.Category, ttl time.Duration) error
	SetAllCategories(ctx context.Context, categories []*domains.Category, ttl time.Duration) error
	DeleteCategory(ctx context.Context, id int) error
}

type cachedCategoryRepository struct {
	baseRepository cache.CachedRepositoryInterface[domains.Category]
}

func NewCachedCategoryRepository(client *redis.Client) CachedCategoryRepository {
	return &cachedCategoryRepository{
		baseRepository: cache.NewBaseCachedRepository[domains.Category](client, "category"),
	}
}

func (r *cachedCategoryRepository) GetCategoryByID(ctx context.Context, id int) (*domains.Category, error) {
	return r.baseRepository.GetByID(ctx, id)
}

func (r *cachedCategoryRepository) GetAllCategories(ctx context.Context) ([]*domains.Category, error) {
	return r.baseRepository.GetAll(ctx)
}

func (r *cachedCategoryRepository) SetCategory(ctx context.Context, category *domains.Category, ttl time.Duration) error {
	return r.baseRepository.Set(ctx, category.ID, category, ttl)
}

func (r *cachedCategoryRepository) SetAllCategories(ctx context.Context, categories []*domains.Category, ttl time.Duration) error {
	return r.baseRepository.SetAll(ctx, categories, ttl)
}

func (r *cachedCategoryRepository) DeleteCategory(ctx context.Context, id int) error {
	return r.baseRepository.Delete(ctx, id)
}
