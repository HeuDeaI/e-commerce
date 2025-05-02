package skintype

import (
	"context"
	"e-commerce/internal/domains"
)

type SkinTypeService interface {
	CreateSkinType(ctx context.Context, skinType *domains.SkinType) (*domains.SkinType, error)
	GetSkinTypeByID(ctx context.Context, id int) (*domains.SkinType, error)
	UpdateSkinType(ctx context.Context, id int, skinType *domains.SkinType) (*domains.SkinType, error)
	DeleteSkinType(ctx context.Context, id int) error
	GetAllSkinTypes(ctx context.Context) ([]*domains.SkinType, error)
}

type skinTypeService struct {
	repo SkinTypeRepository
}

func NewSkinTypeService(repo SkinTypeRepository) SkinTypeService {
	return &skinTypeService{repo: repo}
}

func (s *skinTypeService) CreateSkinType(ctx context.Context, skinType *domains.SkinType) (*domains.SkinType, error) {
	return s.repo.Create(ctx, skinType)
}

func (s *skinTypeService) GetSkinTypeByID(ctx context.Context, id int) (*domains.SkinType, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *skinTypeService) UpdateSkinType(ctx context.Context, id int, skinType *domains.SkinType) (*domains.SkinType, error) {
	return s.repo.Update(ctx, id, skinType)
}

func (s *skinTypeService) DeleteSkinType(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *skinTypeService) GetAllSkinTypes(ctx context.Context) ([]*domains.SkinType, error) {
	return s.repo.GetAll(ctx)
}
