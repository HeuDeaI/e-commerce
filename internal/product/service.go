package product

import (
	"context"
	"e-commerce/internal/domains"
)

type ProductService interface {
	CreateProduct(ctx context.Context, req *domains.ProductRequest) (*domains.ProductResponse, error)
	GetProductByID(ctx context.Context, id int) (*domains.ProductResponse, error)
	UpdateProduct(ctx context.Context, id int, req *domains.ProductRequest) (*domains.ProductResponse, error)
	DeleteProduct(ctx context.Context, id int) error
	GetAllProducts(ctx context.Context) ([]*domains.ProductResponse, error)
	GetProductsByFilter(ctx context.Context, skinTypeIDs []int, brandIDs []int, categoryIDs []int) ([]*domains.ProductResponse, error)
}

type productService struct {
	repo ProductRepository
}

func NewProductService(repo ProductRepository) ProductService {
	return &productService{repo: repo}
}

func (s *productService) CreateProduct(ctx context.Context, req *domains.ProductRequest) (*domains.ProductResponse, error) {
	return s.repo.Create(ctx, req)
}

func (s *productService) GetProductByID(ctx context.Context, id int) (*domains.ProductResponse, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *productService) UpdateProduct(ctx context.Context, id int, req *domains.ProductRequest) (*domains.ProductResponse, error) {
	return s.repo.Update(ctx, id, req)
}

func (s *productService) DeleteProduct(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *productService) GetAllProducts(ctx context.Context) ([]*domains.ProductResponse, error) {
	return s.repo.GetAll(ctx)
}

func (s *productService) GetProductsByFilter(ctx context.Context, skinTypeIDs []int, brandIDs []int, categoryIDs []int) ([]*domains.ProductResponse, error) {
	return s.repo.GetByFilter(ctx, skinTypeIDs, brandIDs, categoryIDs)
}
