package skintype

import (
	"context"
	"database/sql"
	"e-commerce/internal/cache"
	"e-commerce/internal/domains"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type SkinTypeRepository interface {
	Create(ctx context.Context, skinType *domains.SkinType) (*domains.SkinType, error)
	GetByID(ctx context.Context, id int) (*domains.SkinType, error)
	Update(ctx context.Context, id int, skinType *domains.SkinType) (*domains.SkinType, error)
	Delete(ctx context.Context, id int) error
	GetAll(ctx context.Context) ([]*domains.SkinType, error)
}

type skinTypeRepository struct {
	db    *pgxpool.Pool
	cache cache.CacheRepository[domains.SkinType]
}

func NewSkinTypeRepository(db *pgxpool.Pool, redisClient *redis.Client) SkinTypeRepository {
	return &skinTypeRepository{
		db:    db,
		cache: cache.NewCacheRepository[domains.SkinType](redisClient, "skintype"),
	}
}

func (r *skinTypeRepository) Create(ctx context.Context, skinType *domains.SkinType) (*domains.SkinType, error) {
	const insertQuery = `
        INSERT INTO skin_types (name, description)
        VALUES ($1, $2)
        RETURNING id, name, description`

	createdSkin := &domains.SkinType{}
	err := r.db.QueryRow(ctx, insertQuery, skinType.Name, skinType.Description).Scan(
		&createdSkin.ID,
		&createdSkin.Name,
		&createdSkin.Description,
	)
	if err != nil {
		logrus.WithError(err).WithField("skin_type", skinType).Error("Failed to insert skin type")
		return nil, err
	}

	logrus.Infof("Skin type created successfully (ID: %d)", createdSkin.ID)

	if err := r.cache.DeleteAll(ctx); err != nil {
		logrus.Warnf("Failed to clear skin type cache after creation (ID: %d): %v", createdSkin.ID, err)
	}
	go func(s *domains.SkinType) {
		if err := r.cache.Set(context.Background(), s.ID, s); err != nil {
			logrus.Warnf("Failed to cache created skin type asynchronously (ID: %d): %v", s.ID, err)
		} else {
			logrus.Debugf("Successfully cached created skin type asynchronously (ID: %d)", s.ID)
		}
	}(createdSkin)

	return createdSkin, nil
}

func (r *skinTypeRepository) GetByID(ctx context.Context, id int) (*domains.SkinType, error) {
	skin, err := r.cache.GetByID(ctx, id)
	if err == nil {
		logrus.Debugf("Cache hit for skin type (ID: %d)", id)
		return skin, nil
	}
	if !errors.Is(err, redis.Nil) {
		logrus.Warnf("Cache lookup failed for skin type (ID: %d): %v", id, err)
	}

	const getQuery = `SELECT id, name, description FROM skin_types WHERE id = $1`
	skin = &domains.SkinType{}
	err = r.db.QueryRow(ctx, getQuery, id).Scan(
		&skin.ID,
		&skin.Name,
		&skin.Description,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logrus.Warnf("Skin type not found (ID: %d)", id)
			return nil, sql.ErrNoRows
		}
		logrus.Errorf("Failed to get skin type (ID: %d): %v", id, err)
		return nil, err
	}

	logrus.Infof("Skin type received successfuly (ID: %d)", skin.ID)

	go func(s *domains.SkinType) {
		if err := r.cache.Set(context.Background(), s.ID, s); err != nil {
			logrus.Warnf("Failed to cache skin type asynchronously (ID: %d): %v", s.ID, err)
		} else {
			logrus.Debugf("Successfully cached skin type asynchronously (ID: %d)", s.ID)
		}
	}(skin)

	return skin, nil
}

func (r *skinTypeRepository) Update(ctx context.Context, id int, skinType *domains.SkinType) (*domains.SkinType, error) {
	const updateQuery = `
        UPDATE skin_types 
        SET name = $1, description = $2 
        WHERE id = $3
        RETURNING id, name, description`

	updatedSkin := &domains.SkinType{}
	err := r.db.QueryRow(ctx, updateQuery,
		skinType.Name,
		skinType.Description,
		id,
	).Scan(
		&updatedSkin.ID,
		&updatedSkin.Name,
		&updatedSkin.Description,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logrus.Warnf("Attempted to update non-existent skin type (ID: %d)", id)
			return nil, sql.ErrNoRows
		}
		logrus.Errorf("Failed to update skin type (ID: %d): %v", id, err)
		return nil, err
	}

	logrus.Infof("Skin type updated successfully (ID: %d)", updatedSkin.ID)
	if err := r.cache.DeleteAll(ctx); err != nil {
		logrus.Warnf("Failed to clear skin type cache after update (ID: %d): %v", id, err)
	}
	go func(s *domains.SkinType) {
		if err := r.cache.Set(context.Background(), s.ID, s); err != nil {
			logrus.Warnf("Failed to cache updated skin type asynchronously (ID: %d): %v", s.ID, err)
		} else {
			logrus.Debugf("Successfully cached updated skin type asynchronously (ID: %d)", s.ID)
		}
	}(updatedSkin)

	return updatedSkin, nil
}

func (r *skinTypeRepository) Delete(ctx context.Context, id int) error {
	const deleteQuery = `DELETE FROM skin_types WHERE id = $1 RETURNING id`

	var deletedID int
	err := r.db.QueryRow(ctx, deleteQuery, id).Scan(&deletedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logrus.Warnf("Attempted to delete non-existent skin type (ID: %d)", id)
			return sql.ErrNoRows
		}
		logrus.Errorf("Failed to delete skin type (ID: %d): %v", id, err)
		return err
	}

	logrus.Infof("Skin type deleted successfully (ID: %d)", deletedID)

	if err := r.cache.Delete(ctx, id); err != nil {
		logrus.Warnf("Failed to remove skin type from cache (ID: %d): %v", id, err)
	}
	if err := r.cache.DeleteAll(ctx); err != nil {
		logrus.Warnf("Failed to clear all skin types cache after deletion (ID: %d): %v", id, err)
	}

	return nil
}

func (r *skinTypeRepository) GetAll(ctx context.Context) ([]*domains.SkinType, error) {
	skinTypes, cacheErr := r.cache.GetAll(ctx)
	if cacheErr == nil {
		logrus.Debug("Cache hit for all skin types")
		return skinTypes, nil
	}

	const getAllQuery = `SELECT id, name, description FROM skin_types`
	rows, err := r.db.Query(ctx, getAllQuery)
	if err != nil {
		logrus.Errorf("Failed to get all skin types: %v", err)
		return nil, err
	}
	defer rows.Close()

	var skinTypeList []*domains.SkinType
	for rows.Next() {
		skin := &domains.SkinType{}
		if err = rows.Scan(
			&skin.ID,
			&skin.Name,
			&skin.Description,
		); err != nil {
			logrus.Errorf("Failed to scan skin type record: %v", err)
			return nil, err
		}
		skinTypeList = append(skinTypeList, skin)
	}
	if err = rows.Err(); err != nil {
		logrus.Errorf("Error occurred during iteration of rows: %v", err)
		return nil, err
	}

	logrus.Infof("All skin types received successfuly (Count: %d)", len(skinTypeList))

	go func(stList []*domains.SkinType) {
		if err := r.cache.SetAll(context.Background(), stList); err != nil {
			logrus.Warnf("Failed to cache all skin types asynchronously: %v", err)
		} else {
			logrus.Debugf("Successfully cached all skin types asynchronously (Count: %d)", len(stList))
		}
	}(skinTypeList)

	return skinTypeList, nil
}
