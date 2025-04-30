package brand

import (
	"context"
	"e-commerce/internal/domains"
)

type BrandService interface {
	CreateBrand(ctx context.Context, brand *domains.Brand) (*domains.Brand, error)
	GetBrandByID(ctx context.Context, id int) (*domains.Brand, error)
	UpdateBrand(ctx context.Context, id int, brand *domains.Brand) (*domains.Brand, error)
	DeleteBrand(ctx context.Context, id int) error
	GetAllBrands(ctx context.Context) ([]*domains.Brand, error)
}

type brandService struct {
	repo BrandRepository
}

func NewBrandService(repo BrandRepository) BrandService {
	return &brandService{repo: repo}
}

func (s *brandService) CreateBrand(ctx context.Context, brand *domains.Brand) (*domains.Brand, error) {
	return s.repo.Create(ctx, brand)
}

func (s *brandService) GetBrandByID(ctx context.Context, id int) (*domains.Brand, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *brandService) UpdateBrand(ctx context.Context, id int, brand *domains.Brand) (*domains.Brand, error) {
	return s.repo.Update(ctx, id, brand)
}

func (s *brandService) DeleteBrand(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *brandService) GetAllBrands(ctx context.Context) ([]*domains.Brand, error) {
	return s.repo.GetAll(ctx)
}
