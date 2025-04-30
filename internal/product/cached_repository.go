// internal/product/cache.go
package product

import (
	"context"
	"time"

	"e-commerce/internal/cache"
	"e-commerce/internal/domains"
	"github.com/redis/go-redis/v9"
)

type CachedProductRepository interface {
	GetProductByID(ctx context.Context, id int) (*domains.Product, error)
	GetAllProducts(ctx context.Context) ([]*domains.Product, error)
	SetProduct(ctx context.Context, product *domains.Product, ttl time.Duration) error
	SetAllProducts(ctx context.Context, products []*domains.Product, ttl time.Duration) error
	DeleteProduct(ctx context.Context, id int) error
}

type cachedProductRepository struct {
	baseRepository cache.CachedRepositoryInterface[domains.Product]
}

func NewCachedProductRepository(client *redis.Client) CachedProductRepository {
	return &cachedProductRepository{
		baseRepository: cache.NewBaseCachedRepository[domains.Product](client, "product"),
	}
}

func (r *cachedProductRepository) GetProductByID(ctx context.Context, id int) (*domains.Product, error) {
	return r.baseRepository.GetByID(ctx, id)
}

func (r *cachedProductRepository) GetAllProducts(ctx context.Context) ([]*domains.Product, error) {
	return r.baseRepository.GetAll(ctx)
}

func (r *cachedProductRepository) SetProduct(ctx context.Context, product *domains.Product, ttl time.Duration) error {
	return r.baseRepository.Set(ctx, product.ID, product, ttl)
}

func (r *cachedProductRepository) SetAllProducts(ctx context.Context, products []*domains.Product, ttl time.Duration) error {
	return r.baseRepository.SetAll(ctx, products, ttl)
}

func (r *cachedProductRepository) DeleteProduct(ctx context.Context, id int) error {
	return r.baseRepository.Delete(ctx, id)
}
