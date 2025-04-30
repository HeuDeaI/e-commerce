package brand

import (
	"context"
	"time"

	"e-commerce/internal/cache"
	"e-commerce/internal/domains"
	"github.com/redis/go-redis/v9"
)

type CachedBrandRepository interface {
	GetBrandByID(ctx context.Context, id int) (*domains.Brand, error)
	GetAllBrands(ctx context.Context) ([]*domains.Brand, error)
	SetBrand(ctx context.Context, brand *domains.Brand, ttl time.Duration) error
	DeleteBrand(ctx context.Context, id int) error
	SetAllBrands(ctx context.Context, brands []*domains.Brand, ttl time.Duration) error
}

type cachedBrandRepository struct {
	baseRepository cache.CachedRepositoryInterface[domains.Brand]
}

func NewCachedBrandRepository(client *redis.Client) CachedBrandRepository {
	return &cachedBrandRepository{
		baseRepository: cache.NewBaseCachedRepository[domains.Brand](client, "brand"),
	}
}

func (r *cachedBrandRepository) GetBrandByID(ctx context.Context, id int) (*domains.Brand, error) {
	return r.baseRepository.GetByID(ctx, id)
}

func (r *cachedBrandRepository) GetAllBrands(ctx context.Context) ([]*domains.Brand, error) {
	return r.baseRepository.GetAll(ctx)
}

func (r *cachedBrandRepository) SetBrand(ctx context.Context, brand *domains.Brand, ttl time.Duration) error {
	return r.baseRepository.Set(ctx, brand.ID, brand, ttl)
}

func (r *cachedBrandRepository) SetAllBrands(ctx context.Context, brands []*domains.Brand, ttl time.Duration) error {
	return r.baseRepository.SetAll(ctx, brands, ttl)
}

func (r *cachedBrandRepository) DeleteBrand(ctx context.Context, id int) error {
	return r.baseRepository.Delete(ctx, id)
}
