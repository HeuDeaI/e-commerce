package product

import (
	"context"
	"e-commerce/internal/domains"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type ProductService interface {
	CreateProduct(ctx context.Context, product *domains.Product) (*domains.Product, error)
	GetProductByID(ctx context.Context, id uint) (*domains.Product, error)
	UpdateProduct(ctx context.Context, id uint, product *domains.Product) (*domains.Product, error)
	DeleteProduct(ctx context.Context, id uint) error
	GetAllProducts(ctx context.Context) ([]*domains.Product, error)
}

type productService struct {
	repo  ProductRepository
	cache CachedProductRepository
}

func NewProductService(repo ProductRepository, cache CachedProductRepository) ProductService {
	return &productService{
		repo:  repo,
		cache: cache,
	}
}

func (s *productService) CreateProduct(ctx context.Context, product *domains.Product) (*domains.Product, error) {
	createdProduct, err := s.repo.CreateProduct(ctx, product)
	if err != nil {
		return nil, err
	}

	if err := s.cache.SetProduct(ctx, createdProduct, 10*time.Minute); err != nil {
		return createdProduct, err
	}

	return createdProduct, nil
}

func (s *productService) GetProductByID(ctx context.Context, id uint) (*domains.Product, error) {
	product, err := s.cache.GetProductByID(ctx, id)
	if err == nil {
		return product, nil
	} else if err != redis.Nil {
		return nil, err
	}

	product, err = s.repo.GetProductByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.cache.SetProduct(ctx, product, 10*time.Minute); err != nil {
		return product, err
	}
	return product, nil
}

func (s *productService) UpdateProduct(ctx context.Context, id uint, product *domains.Product) (*domains.Product, error) {
	updatedProduct, err := s.repo.UpdateProduct(ctx, id, product)
	if err != nil {
		return nil, err
	}

	if err := s.cache.SetProduct(ctx, updatedProduct, 10*time.Minute); err != nil {
		return updatedProduct, err
	}

	return updatedProduct, nil
}

func (s *productService) DeleteProduct(ctx context.Context, id uint) error {
	err := s.repo.DeleteProduct(ctx, id)
	if err != nil {
		return err
	}

	if err := s.cache.DeleteCache(ctx, fmt.Sprintf("product:%d", id)); err != nil {
		return err
	}

	return nil
}

func (s *productService) GetAllProducts(ctx context.Context) ([]*domains.Product, error) {
	products, err := s.cache.GetAllProducts(ctx)
	if err == nil {
		return products, nil
	} else if err != redis.Nil {
		return nil, err
	}

	products, err = s.repo.GetAllProducts(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.cache.SetAllProducts(ctx, products, 10*time.Minute); err != nil {
		return products, err
	}
	return products, nil
}
