package brand

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

type BrandRepository interface {
	Create(ctx context.Context, brand *domains.Brand) (*domains.Brand, error)
	GetByID(ctx context.Context, id int) (*domains.Brand, error)
	Update(ctx context.Context, id int, brand *domains.Brand) (*domains.Brand, error)
	Delete(ctx context.Context, id int) error
	GetAll(ctx context.Context) ([]*domains.Brand, error)
}

type brandRepository struct {
	db    *pgxpool.Pool
	cache cache.CacheRepository[domains.Brand]
}

func NewBrandRepository(db *pgxpool.Pool, redisClient *redis.Client) BrandRepository {
	return &brandRepository{
		db:    db,
		cache: cache.NewCacheRepository[domains.Brand](redisClient, "brand"),
	}
}

func (r *brandRepository) Create(ctx context.Context, brand *domains.Brand) (*domains.Brand, error) {
	const insertQuery = `
        INSERT INTO brands (name, description, website)
        VALUES ($1, $2, $3)
        RETURNING id, name, description, website`

	createdBrand := &domains.Brand{}
	err := r.db.QueryRow(ctx, insertQuery, brand.Name, brand.Description, brand.Website).Scan(
		&createdBrand.ID,
		&createdBrand.Name,
		&createdBrand.Description,
		&createdBrand.Website,
	)
	if err != nil {
		logrus.WithError(err).WithField("brand", brand).Error("Failed to insert brand")
		return nil, err
	}

	logrus.Debugf("Brand created successfully (ID: %d)", createdBrand.ID)

	if err := r.cache.DeleteAll(ctx); err != nil {
		logrus.Warnf("Failed to clear brand cache after creation (ID: %d): %v", createdBrand.ID, err)
	}
	go func(b *domains.Brand) {
		if err := r.cache.Set(context.Background(), b.ID, b); err != nil {
			logrus.Warnf("Failed to cache created brand asynchronously (ID: %d): %v", b.ID, err)
		} else {
			logrus.Debugf("Successfully cached created brand asynchronously (ID: %d)", b.ID)
		}
	}(createdBrand)

	return createdBrand, nil
}

func (r *brandRepository) GetByID(ctx context.Context, id int) (*domains.Brand, error) {
	brand, err := r.cache.GetByID(ctx, id)
	if err == nil {
		logrus.Debugf("Cache hit for brand (ID: %d)", id)
		return brand, nil
	}
	if !errors.Is(err, redis.Nil) {
		logrus.Errorf("Cache lookup failed for brand (ID: %d): %v", id, err)
	}

	const getQuery = `SELECT id, name, description, website FROM brands WHERE id = $1`
	brand = &domains.Brand{}
	err = r.db.QueryRow(ctx, getQuery, id).Scan(
		&brand.ID,
		&brand.Name,
		&brand.Description,
		&brand.Website,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logrus.Infof("Brand not found (ID: %d)", id)
			return nil, sql.ErrNoRows
		}
		logrus.Errorf("Failed to get brand (ID: %d): %v", id, err)
		return nil, err
	}

	logrus.Debugf("Brand retrieved successfully (ID: %d)", brand.ID)

	go func(b *domains.Brand) {
		if err := r.cache.Set(context.Background(), b.ID, b); err != nil {
			logrus.Warnf("Failed to cache brand asynchronously (ID: %d): %v", b.ID, err)
		} else {
			logrus.Debugf("Successfully cached brand asynchronously (ID: %d)", b.ID)
		}
	}(brand)

	return brand, nil
}

func (r *brandRepository) Update(ctx context.Context, id int, brand *domains.Brand) (*domains.Brand, error) {
	const updateQuery = `
        UPDATE brands 
        SET name = $1, description = $2, website = $3 
        WHERE id = $4
        RETURNING id, name, description, website`

	updatedBrand := &domains.Brand{}
	err := r.db.QueryRow(ctx, updateQuery,
		brand.Name,
		brand.Description,
		brand.Website,
		id,
	).Scan(
		&updatedBrand.ID,
		&updatedBrand.Name,
		&updatedBrand.Description,
		&updatedBrand.Website,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logrus.Infof("Attempted to update non-existent brand (ID: %d)", id)
			return nil, sql.ErrNoRows
		}
		logrus.Errorf("Failed to update brand (ID: %d): %v", id, err)
		return nil, err
	}

	logrus.Debugf("Brand updated successfully (ID: %d)", updatedBrand.ID)

	if err := r.cache.DeleteAll(ctx); err != nil {
		logrus.Warnf("Failed to clear brand cache after update (ID: %d): %v", id, err)
	}
	go func(b *domains.Brand) {
		if err := r.cache.Set(context.Background(), b.ID, b); err != nil {
			logrus.Warnf("Failed to cache updated brand asynchronously (ID: %d): %v", b.ID, err)
		} else {
			logrus.Debugf("Successfully cached updated brand asynchronously (ID: %d)", b.ID)
		}
	}(updatedBrand)

	return updatedBrand, nil
}

func (r *brandRepository) Delete(ctx context.Context, id int) error {
	const deleteQuery = `DELETE FROM brands WHERE id = $1 RETURNING id`

	var deletedID int
	err := r.db.QueryRow(ctx, deleteQuery, id).Scan(&deletedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logrus.Infof("Attempted to delete non-existent brand (ID: %d)", id)
			return sql.ErrNoRows
		}
		logrus.Errorf("Failed to delete brand (ID: %d): %v", id, err)
		return err
	}

	logrus.Debugf("Brand deleted successfully (ID: %d)", deletedID)

	if err := r.cache.Delete(ctx, id); err != nil {
		logrus.Warnf("Failed to remove brand from cache (ID: %d): %v", id, err)
	}
	if err := r.cache.DeleteAll(ctx); err != nil {
		logrus.Warnf("Failed to clear all brands cache after deletion (ID: %d): %v", id, err)
	}

	return nil
}

func (r *brandRepository) GetAll(ctx context.Context) ([]*domains.Brand, error) {
	brands, err := r.cache.GetAll(ctx)
	if err == nil {
		logrus.Debug("Cache hit for all brands")
		return brands, nil
	}
	if !errors.Is(err, redis.Nil) {
		logrus.Errorf("Cache lookup failed for all brands: %v", err)
	}

	const getAllQuery = `SELECT id, name, description, website FROM brands`
	rows, err := r.db.Query(ctx, getAllQuery)
	if err != nil {
		logrus.Errorf("Failed to get all brands: %v", err)
		return nil, err
	}
	defer rows.Close()

	var brandsList []*domains.Brand
	for rows.Next() {
		brand := &domains.Brand{}
		if err = rows.Scan(
			&brand.ID,
			&brand.Name,
			&brand.Description,
			&brand.Website,
		); err != nil {
			logrus.Errorf("Failed to scan brand record: %v", err)
			return nil, err
		}
		brandsList = append(brandsList, brand)
	}
	if rows.Err() != nil {
		logrus.Errorf("Error occurred during iteration of rows: %v", rows.Err())
		return nil, rows.Err()
	}

	logrus.Debugf("All brands retrieved successfully (Count: %d)", len(brandsList))

	go func(bl []*domains.Brand) {
		if err := r.cache.SetAll(context.Background(), bl); err != nil {
			logrus.Warnf("Failed to cache all brands asynchronously: %v", err)
		} else {
			logrus.Debugf("Successfully cached all brands asynchronously (Count: %d)", len(bl))
		}
	}(brandsList)

	return brandsList, nil
}
