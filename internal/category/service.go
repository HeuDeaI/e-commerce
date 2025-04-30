package category

import (
	"context"
	"e-commerce/internal/domains"
	"time"

	"github.com/redis/go-redis/v9"
)

type CategoryService interface {
	CreateCategory(ctx context.Context, category *domains.Category) (*domains.Category, error)
	GetCategoryByID(ctx context.Context, id int) (*domains.Category, error)
	UpdateCategory(ctx context.Context, id int, category *domains.Category) (*domains.Category, error)
	DeleteCategory(ctx context.Context, id int) error
	GetAllCategories(ctx context.Context) ([]*domains.Category, error)
}

type categoryService struct {
	repo  CategoryRepository
	cache CachedCategoryRepository
}

func NewCategoryService(repo CategoryRepository, cache CachedCategoryRepository) CategoryService {
	return &categoryService{
		repo:  repo,
		cache: cache,
	}
}

func (s *categoryService) CreateCategory(ctx context.Context, category *domains.Category) (*domains.Category, error) {
	createdCategory, err := s.repo.CreateCategory(ctx, category)
	if err != nil {
		return nil, err
	}

	if err := s.cache.SetCategory(ctx, createdCategory, 10*time.Minute); err != nil {
		return createdCategory, err
	}

	return createdCategory, nil
}

func (s *categoryService) GetCategoryByID(ctx context.Context, id int) (*domains.Category, error) {
	category, err := s.cache.GetCategoryByID(ctx, id)
	if err == nil {
		return category, nil
	} else if err != redis.Nil {
		return nil, err
	}

	category, err = s.repo.GetCategoryByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.cache.SetCategory(ctx, category, 10*time.Minute); err != nil {
		return category, err
	}

	return category, nil
}

func (s *categoryService) UpdateCategory(ctx context.Context, id int, category *domains.Category) (*domains.Category, error) {
	updatedCategory, err := s.repo.UpdateCategory(ctx, id, category)
	if err != nil {
		return nil, err
	}

	if err := s.cache.SetCategory(ctx, updatedCategory, 10*time.Minute); err != nil {
		return updatedCategory, err
	}

	return updatedCategory, nil
}

func (s *categoryService) DeleteCategory(ctx context.Context, id int) error {
	if err := s.repo.DeleteCategory(ctx, id); err != nil {
		return err
	}

	return s.cache.DeleteCategory(ctx, id)
}

func (s *categoryService) GetAllCategories(ctx context.Context) ([]*domains.Category, error) {
	categories, err := s.cache.GetAllCategories(ctx)
	if err == nil {
		return categories, nil
	} else if err != redis.Nil {
		return nil, err
	}

	categories, err = s.repo.GetAllCategories(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.cache.SetAllCategories(ctx, categories, 10*time.Minute); err != nil {
		return categories, err
	}

	return categories, nil
}
