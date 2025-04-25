package brand

import (
	"context"
	"e-commerce/internal/domains"
)

type BrandService interface {
	CreateBrand(ctx context.Context, brand *domains.Brand) (*domains.Brand, error)
	GetBrandByID(ctx context.Context, id uint) (*domains.Brand, error)
	UpdateBrand(ctx context.Context, id uint, brand *domains.Brand) (*domains.Brand, error)
	DeleteBrand(ctx context.Context, id uint) error
	GetAllBrands(ctx context.Context) ([]*domains.Brand, error)
}

type brandService struct {
	repo BrandRepository
}

func NewBrandService(repo BrandRepository) BrandService {
	return &brandService{repo: repo}
}

func (s *brandService) CreateBrand(ctx context.Context, brand *domains.Brand) (*domains.Brand, error) {
	return s.repo.CreateBrand(ctx, brand)
}

func (s *brandService) GetBrandByID(ctx context.Context, id uint) (*domains.Brand, error) {
	return s.repo.GetBrandByID(ctx, id)
}

func (s *brandService) UpdateBrand(ctx context.Context, id uint, brand *domains.Brand) (*domains.Brand, error) {
	return s.repo.UpdateBrand(ctx, id, brand)
}

func (s *brandService) DeleteBrand(ctx context.Context, id uint) error {
	return s.repo.DeleteBrand(ctx, id)
}

func (s *brandService) GetAllBrands(ctx context.Context) ([]*domains.Brand, error) {
	return s.repo.GetAllBrands(ctx)
}
