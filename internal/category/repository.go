package category

import (
	"context"
	"e-commerce/internal/cache"
	"e-commerce/internal/domains"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type CategoryRepository interface {
	Create(ctx context.Context, category *domains.Category) (*domains.Category, error)
	GetByID(ctx context.Context, id int) (*domains.Category, error)
	Update(ctx context.Context, id int, category *domains.Category) (*domains.Category, error)
	Delete(ctx context.Context, id int) error
	GetAll(ctx context.Context) ([]*domains.Category, error)
}

type categoryRepository struct {
	db    *pgxpool.Pool
	cache cache.CachedRepositoryInterface[domains.Category]
	ttl   time.Duration
}

func NewCategoryRepository(db *pgxpool.Pool, redisClient *redis.Client, ttl time.Duration) CategoryRepository {
	return &categoryRepository{
		db:    db,
		cache: cache.NewBaseCachedRepository[domains.Category](redisClient, "category"),
		ttl:   ttl,
	}
}

func (r *categoryRepository) Create(ctx context.Context, category *domains.Category) (*domains.Category, error) {
	insertQuery := `
        INSERT INTO categories (name, description) 
        VALUES ($1, $2) 
        RETURNING id`

	var id int
	err := r.db.QueryRow(ctx, insertQuery, category.Name, category.Description).Scan(&id)
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

func (r *categoryRepository) GetByID(ctx context.Context, id int) (*domains.Category, error) {
	if category, err := r.cache.GetByID(ctx, id); err == nil {
		return category, nil
	}

	query := `
        SELECT id, name, description 
        FROM categories 
        WHERE id = $1`

	row := r.db.QueryRow(ctx, query, id)
	category, err := scanCategoryRow(row)
	if err != nil {
		return nil, err
	}

	_ = r.cache.Set(ctx, category.ID, category, r.ttl)
	return category, nil
}

func scanCategoryRow(row pgx.Row) (*domains.Category, error) {
	category := &domains.Category{}
	err := row.Scan(&category.ID, &category.Name, &category.Description)
	return category, err
}

func (r *categoryRepository) Update(ctx context.Context, id int, category *domains.Category) (*domains.Category, error) {
	updateQuery := `
        UPDATE categories 
        SET name = $1, description = $2 
        WHERE id = $3`

	_, err := r.db.Exec(ctx, updateQuery, category.Name, category.Description, id)
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

func (r *categoryRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM categories WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	_ = r.cache.Delete(ctx, id)
	_ = r.cache.DeleteAll(ctx)
	return nil
}

func (r *categoryRepository) GetAll(ctx context.Context) ([]*domains.Category, error) {
	if categories, err := r.cache.GetAll(ctx); err == nil {
		return categories, nil
	}

	query := `
        SELECT id, name, description 
        FROM categories`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*domains.Category
	for rows.Next() {
		category := &domains.Category{}
		err := rows.Scan(&category.ID, &category.Name, &category.Description)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	_ = r.cache.SetAll(ctx, categories, r.ttl)
	return categories, nil
}
