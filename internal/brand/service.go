package brand

import (
	"context"
	"e-commerce/internal/domains"
	"time"

	"github.com/redis/go-redis/v9"
)

type BrandService interface {
	CreateBrand(ctx context.Context, brand *domains.Brand) (*domains.Brand, error)
	GetBrandByID(ctx context.Context, id int) (*domains.Brand, error)
	UpdateBrand(ctx context.Context, id int, brand *domains.Brand) (*domains.Brand, error)
	DeleteBrand(ctx context.Context, id int) error
	GetAllBrands(ctx context.Context) ([]*domains.Brand, error)
}

type brandService struct {
	repo  BrandRepository
	cache CachedBrandRepository
}

func NewBrandService(repo BrandRepository, cache CachedBrandRepository) BrandService {
	return &brandService{repo: repo, cache: cache}
}

func (s *brandService) CreateBrand(ctx context.Context, brand *domains.Brand) (*domains.Brand, error) {
	createdBrand, err := s.repo.CreateBrand(ctx, brand)
	if err != nil {
		return nil, err
	}

	if err := s.cache.SetBrand(ctx, createdBrand, 10*time.Minute); err != nil {
		return createdBrand, err
	}

	return createdBrand, nil
}

func (s *brandService) GetBrandByID(ctx context.Context, id int) (*domains.Brand, error) {
	brand, err := s.cache.GetBrandByID(ctx, id)
	if err == nil {
		return brand, nil
	} else if err != redis.Nil {
		return nil, err
	}

	brand, err = s.repo.GetBrandByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.cache.SetBrand(ctx, brand, 10*time.Minute); err != nil {
		return brand, err
	}

	return brand, nil
}

func (s *brandService) UpdateBrand(ctx context.Context, id int, brand *domains.Brand) (*domains.Brand, error) {
	updatedBrand, err := s.repo.UpdateBrand(ctx, id, brand)
	if err != nil {
		return nil, err
	}

	if err := s.cache.SetBrand(ctx, updatedBrand, 10*time.Minute); err != nil {
		return updatedBrand, err
	}

	return updatedBrand, nil
}

func (s *brandService) DeleteBrand(ctx context.Context, id int) error {
	if err := s.repo.DeleteBrand(ctx, id); err != nil {
		return err
	}

	return s.cache.DeleteBrand(ctx, id)
}

func (s *brandService) GetAllBrands(ctx context.Context) ([]*domains.Brand, error) {
	brands, err := s.cache.GetAllBrands(ctx)
	if err == nil {
		return brands, nil
	} else if err != redis.Nil {
		return nil, err
	}

	brands, err = s.repo.GetAllBrands(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.cache.SetAllBrands(ctx, brands, 10*time.Minute); err != nil {
		return brands, err
	}

	return brands, nil
}
