package brand

import (
	"context"
	"e-commerce/internal/cache"
	"e-commerce/internal/domains"

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
	insertQuery := `
        INSERT INTO brands (name, description, website)
        VALUES ($1, $2, $3) RETURNING id`

	var id int
	err := r.db.QueryRow(ctx, insertQuery, brand.Name, brand.Description, brand.Website).Scan(&id)
	if err != nil {
		return nil, err
	}

	created, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err = r.cache.DeleteAll(ctx); err != nil {
		return created, err
	}
	if err = r.cache.Set(ctx, created.ID, created); err != nil {
		return created, err
	}

	return created, nil
}

func (r *brandRepository) GetByID(ctx context.Context, id int) (*domains.Brand, error) {
	brand, err := r.cache.GetByID(ctx, id)
	if err == nil {
		logrus.WithFields(logrus.Fields{"id": id}).Info("Cache hit for brand")
		return brand, nil
	}
	if err != redis.Nil {
		logrus.WithFields(logrus.Fields{
			"id":    id,
			"error": err.Error(),
		}).Error("Cache lookup failed for brand")
	}

	query := `SELECT id, name, description, website FROM brands WHERE id = $1`
	brand = &domains.Brand{}
	err = r.db.QueryRow(ctx, query, id).Scan(&brand.ID, &brand.Name, &brand.Description, &brand.Website)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"id":    id,
			"query": query,
			"error": err.Error(),
		}).Error("Database query failed for brand")
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"id":    brand.ID,
		"query": query,
	}).Info("Database query successful for brand")

	if err := r.cache.Set(ctx, brand.ID, brand); err != nil {
		logrus.WithFields(logrus.Fields{
			"id":    brand.ID,
			"error": err.Error(),
		}).Warn("Failed to cache brand")
	} else {
		logrus.WithFields(logrus.Fields{"id": brand.ID}).Info("Successfully cached brand")
	}

	return brand, nil
}

func (r *brandRepository) Update(ctx context.Context, id int, brand *domains.Brand) (*domains.Brand, error) {
	updateQuery := `
        UPDATE brands 
        SET name = $1, description = $2, website = $3 
        WHERE id = $4`

	_, err := r.db.Exec(ctx, updateQuery, brand.Name, brand.Description, brand.Website, id)
	if err != nil {
		return nil, err
	}

	updated, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err = r.cache.DeleteAll(ctx); err != nil {
		return updated, err
	}
	if err = r.cache.Set(ctx, updated.ID, updated); err != nil {
		return updated, err
	}

	return updated, nil
}

func (r *brandRepository) Delete(ctx context.Context, id int) error {
	deleteQuery := `DELETE FROM brands WHERE id = $1`
	_, err := r.db.Exec(ctx, deleteQuery, id)
	if err != nil {
		return err
	}

	if err = r.cache.Delete(ctx, id); err != nil {
		return err
	}
	if err = r.cache.DeleteAll(ctx); err != nil {
		return err
	}

	return nil
}

func (r *brandRepository) GetAll(ctx context.Context) ([]*domains.Brand, error) {
	brands, cacheErr := r.cache.GetAll(ctx)
	if cacheErr == nil {
		return brands, nil
	}

	query := `
        SELECT id, name, description, website 
        FROM brands`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var brandsList []*domains.Brand
	for rows.Next() {
		brand := &domains.Brand{}
		if err = rows.Scan(&brand.ID, &brand.Name, &brand.Description, &brand.Website); err != nil {
			return nil, err
		}
		brandsList = append(brandsList, brand)
	}

	if cacheErr != redis.Nil {
		return brands, err
	}
	if err = r.cache.SetAll(ctx, brandsList); err != nil {
		return brandsList, err
	}

	return brandsList, nil
}
