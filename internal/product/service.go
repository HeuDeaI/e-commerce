package product

import (
	"context"
	"e-commerce/internal/domains"
)

type ProductService interface {
	CreateProduct(ctx context.Context, product *domains.Product) (*domains.Product, error)
	GetProductByID(ctx context.Context, id int) (*domains.Product, error)
	UpdateProduct(ctx context.Context, id int, product *domains.Product) (*domains.Product, error)
	DeleteProduct(ctx context.Context, id int) error
	GetAllProducts(ctx context.Context) ([]*domains.Product, error)
}

type productService struct {
	repo ProductRepository
}

func NewProductService(repo ProductRepository) ProductService {
	return &productService{repo: repo}
}
func (s *productService) CreateProduct(ctx context.Context, product *domains.Product) (*domains.Product, error) {
	return s.repo.Create(ctx, product)
}

func (s *productService) GetProductByID(ctx context.Context, id int) (*domains.Product, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *productService) UpdateProduct(ctx context.Context, id int, product *domains.Product) (*domains.Product, error) {
	return s.repo.Update(ctx, id, product)
}

func (s *productService) DeleteProduct(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *productService) GetAllProducts(ctx context.Context) ([]*domains.Product, error) {
	return s.repo.GetAll(ctx)
}
