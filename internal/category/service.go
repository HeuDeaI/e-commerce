package category

import (
	"context"
	"e-commerce/internal/domains"
)

type CategoryService interface {
	CreateCategory(ctx context.Context, category *domains.Category) (*domains.Category, error)
	GetCategoryByID(ctx context.Context, id int) (*domains.Category, error)
	UpdateCategory(ctx context.Context, id int, category *domains.Category) (*domains.Category, error)
	DeleteCategory(ctx context.Context, id int) error
	GetAllCategories(ctx context.Context) ([]*domains.Category, error)
}

type categoryService struct {
	repo CategoryRepository
}

func NewCategoryService(repo CategoryRepository) CategoryService {
	return &categoryService{repo: repo}
}

func (s *categoryService) CreateCategory(ctx context.Context, category *domains.Category) (*domains.Category, error) {
	return s.repo.Create(ctx, category)
}

func (s *categoryService) GetCategoryByID(ctx context.Context, id int) (*domains.Category, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *categoryService) UpdateCategory(ctx context.Context, id int, category *domains.Category) (*domains.Category, error) {
	return s.repo.Update(ctx, id, category)
}

func (s *categoryService) DeleteCategory(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *categoryService) GetAllCategories(ctx context.Context) ([]*domains.Category, error) {
	return s.repo.GetAll(ctx)
}
