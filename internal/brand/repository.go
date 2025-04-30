package brand

import (
	"context"
	"e-commerce/internal/cache"
	"e-commerce/internal/domains"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
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
	cache cache.CachedRepositoryInterface[domains.Brand]
	ttl   time.Duration
}

func NewBrandRepository(db *pgxpool.Pool, redisClient *redis.Client, ttl time.Duration) BrandRepository {
	return &brandRepository{
		db:    db,
		cache: cache.NewBaseCachedRepository[domains.Brand](redisClient, "brand"),
		ttl:   ttl,
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

	_ = r.cache.DeleteAll(ctx)
	_ = r.cache.Set(ctx, created.ID, created, r.ttl)
	return created, nil
}

func (r *brandRepository) GetByID(ctx context.Context, id int) (*domains.Brand, error) {
	if brand, err := r.cache.GetByID(ctx, id); err == nil {
		return brand, nil
	}

	query := `
        SELECT id, name, description, website 
        FROM brands 
        WHERE id = $1`

	brand := &domains.Brand{}
	row := r.db.QueryRow(ctx, query, id)
	err := row.Scan(&brand.ID, &brand.Name, &brand.Description, &brand.Website)
	if err != nil {
		return nil, err
	}

	_ = r.cache.Set(ctx, brand.ID, brand, r.ttl)
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

	_ = r.cache.DeleteAll(ctx)
	_ = r.cache.Set(ctx, updated.ID, updated, r.ttl)
	return updated, nil
}

func (r *brandRepository) Delete(ctx context.Context, id int) error {
	deleteQuery := `DELETE FROM brands WHERE id = $1`
	_, err := r.db.Exec(ctx, deleteQuery, id)
	if err != nil {
		return err
	}

	_ = r.cache.Delete(ctx, id)
	_ = r.cache.DeleteAll(ctx)
	return nil
}

func (r *brandRepository) GetAll(ctx context.Context) ([]*domains.Brand, error) {
	if brands, err := r.cache.GetAll(ctx); err == nil {
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

	var brands []*domains.Brand
	for rows.Next() {
		brand := &domains.Brand{}
		err := rows.Scan(&brand.ID, &brand.Name, &brand.Description, &brand.Website)
		if err != nil {
			return nil, err
		}
		brands = append(brands, brand)
	}

	_ = r.cache.SetAll(ctx, brands, r.ttl)
	return brands, nil
}
