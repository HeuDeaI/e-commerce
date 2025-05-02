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

	createdSkinType := &domains.SkinType{}
	err := r.db.QueryRow(ctx, insertQuery, skinType.Name, skinType.Description).Scan(
		&createdSkinType.ID,
		&createdSkinType.Name,
		&createdSkinType.Description,
	)
	if err != nil {
		logrus.WithError(err).WithField("skinType", skinType).Error("Failed to insert skin type")
		return nil, err
	}

	logrus.Debugf("Skin type created successfully (ID: %d)", createdSkinType.ID)

	if err := r.cache.DeleteAll(ctx); err != nil {
		logrus.Warnf("Failed to clear skin type cache after creation (ID: %d): %v", createdSkinType.ID, err)
	}
	go func(st *domains.SkinType) {
		if err := r.cache.Set(context.Background(), st.ID, st); err != nil {
			logrus.Warnf("Failed to cache created skin type asynchronously (ID: %d): %v", st.ID, err)
		} else {
			logrus.Debugf("Successfully cached created skin type asynchronously (ID: %d)", st.ID)
		}
	}(createdSkinType)

	return createdSkinType, nil
}

func (r *skinTypeRepository) GetByID(ctx context.Context, id int) (*domains.SkinType, error) {
	skinType, err := r.cache.GetByID(ctx, id)
	if err == nil {
		logrus.Debugf("Cache hit for skin type (ID: %d)", id)
		return skinType, nil
	}
	if !errors.Is(err, redis.Nil) {
		logrus.Errorf("Cache lookup failed for skin type (ID: %d): %v", id, err)
	}

	const getQuery = `SELECT id, name, description FROM skin_types WHERE id = $1`
	skinType = &domains.SkinType{}
	err = r.db.QueryRow(ctx, getQuery, id).Scan(
		&skinType.ID,
		&skinType.Name,
		&skinType.Description,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logrus.Infof("Skin type not found (ID: %d)", id)
			return nil, sql.ErrNoRows
		}
		logrus.Errorf("Failed to get skin type (ID: %d): %v", id, err)
		return nil, err
	}

	logrus.Debugf("Skin type retrieved successfully (ID: %d)", skinType.ID)

	go func(st *domains.SkinType) {
		if err := r.cache.Set(context.Background(), st.ID, st); err != nil {
			logrus.Warnf("Failed to cache skin type asynchronously (ID: %d): %v", st.ID, err)
		} else {
			logrus.Debugf("Successfully cached skin type asynchronously (ID: %d)", st.ID)
		}
	}(skinType)

	return skinType, nil
}

func (r *skinTypeRepository) Update(ctx context.Context, id int, skinType *domains.SkinType) (*domains.SkinType, error) {
	const updateQuery = `
        UPDATE skin_types 
        SET name = $1, description = $2 
        WHERE id = $3
        RETURNING id, name, description`

	updatedSkinType := &domains.SkinType{}
	err := r.db.QueryRow(ctx, updateQuery,
		skinType.Name,
		skinType.Description,
		id,
	).Scan(
		&updatedSkinType.ID,
		&updatedSkinType.Name,
		&updatedSkinType.Description,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logrus.Infof("Attempted to update non-existent skin type (ID: %d)", id)
			return nil, sql.ErrNoRows
		}
		logrus.Errorf("Failed to update skin type (ID: %d): %v", id, err)
		return nil, err
	}

	logrus.Debugf("Skin type updated successfully (ID: %d)", updatedSkinType.ID)

	if err := r.cache.DeleteAll(ctx); err != nil {
		logrus.Warnf("Failed to clear skin type cache after update (ID: %d): %v", id, err)
	}
	go func(st *domains.SkinType) {
		if err := r.cache.Set(context.Background(), st.ID, st); err != nil {
			logrus.Warnf("Failed to cache updated skin type asynchronously (ID: %d): %v", st.ID, err)
		} else {
			logrus.Debugf("Successfully cached updated skin type asynchronously (ID: %d)", st.ID)
		}
	}(updatedSkinType)

	return updatedSkinType, nil
}

func (r *skinTypeRepository) Delete(ctx context.Context, id int) error {
	const deleteQuery = `DELETE FROM skin_types WHERE id = $1 RETURNING id`

	var deletedID int
	err := r.db.QueryRow(ctx, deleteQuery, id).Scan(&deletedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logrus.Infof("Attempted to delete non-existent skin type (ID: %d)", id)
			return sql.ErrNoRows
		}
		logrus.Errorf("Failed to delete skin type (ID: %d): %v", id, err)
		return err
	}

	logrus.Debugf("Skin type deleted successfully (ID: %d)", deletedID)

	if err := r.cache.Delete(ctx, id); err != nil {
		logrus.Warnf("Failed to remove skin type from cache (ID: %d): %v", id, err)
	}
	if err := r.cache.DeleteAll(ctx); err != nil {
		logrus.Warnf("Failed to clear all skin types cache after deletion (ID: %d): %v", id, err)
	}

	return nil
}

func (r *skinTypeRepository) GetAll(ctx context.Context) ([]*domains.SkinType, error) {
	skinTypes, err := r.cache.GetAll(ctx)
	if err == nil {
		logrus.Debug("Cache hit for all skin types")
		return skinTypes, nil
	}
	if !errors.Is(err, redis.Nil) {
		logrus.Errorf("Cache lookup failed for all skin types: %v", err)
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
		skinType := &domains.SkinType{}
		if err = rows.Scan(
			&skinType.ID,
			&skinType.Name,
			&skinType.Description,
		); err != nil {
			logrus.Errorf("Failed to scan skin type record: %v", err)
			return nil, err
		}
		skinTypeList = append(skinTypeList, skinType)
	}
	if err = rows.Err(); err != nil {
		logrus.Errorf("Error occurred during iteration of rows: %v", err)
		return nil, err
	}

	logrus.Debugf("All skin types retrieved successfully (Count: %d)", len(skinTypeList))

	go func(stList []*domains.SkinType) {
		if err := r.cache.SetAll(context.Background(), stList); err != nil {
			logrus.Warnf("Failed to cache all skin types asynchronously: %v", err)
		} else {
			logrus.Debugf("Successfully cached all skin types asynchronously (Count: %d)", len(stList))
		}
	}(skinTypeList)

	return skinTypeList, nil
}
