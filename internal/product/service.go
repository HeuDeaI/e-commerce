package product

import (
	"context"
	"e-commerce/internal/domains"
)

type ProductService interface {
	CreateProduct(ctx context.Context, product *domains.Product) (*domains.Product, error)
	GetProductByID(ctx context.Context, id uint) (*domains.Product, error)
	UpdateProduct(ctx context.Context, id uint, product *domains.Product) (*domains.Product, error)
	DeleteProduct(ctx context.Context, id uint) error
	GetAllProducts(ctx context.Context) ([]*domains.Product, error)
}

type productService struct {
	repo ProductRepository
}

func NewProductService(repo ProductRepository) ProductService {
	return &productService{repo: repo}
}

func (s *productService) CreateProduct(ctx context.Context, product *domains.Product) (*domains.Product, error) {
	return s.repo.CreateProduct(ctx, product)
}

func (s *productService) GetProductByID(ctx context.Context, id uint) (*domains.Product, error) {
	return s.repo.GetProductByID(ctx, id)
}

func (s *productService) UpdateProduct(ctx context.Context, id uint, product *domains.Product) (*domains.Product, error) {
	return s.repo.UpdateProduct(ctx, id, product)
}

func (s *productService) DeleteProduct(ctx context.Context, id uint) error {
	return s.repo.DeleteProduct(ctx, id)
}

func (s *productService) GetAllProducts(ctx context.Context) ([]*domains.Product, error) {
	return s.repo.GetAllProducts(ctx)
}
