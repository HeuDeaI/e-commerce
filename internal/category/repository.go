package category

import (
	"context"
	"e-commerce/internal/cache"
	"e-commerce/internal/domains"

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
	cache cache.CacheRepository[domains.Category]
}

func NewCategoryRepository(db *pgxpool.Pool, redisClient *redis.Client) CategoryRepository {
	return &categoryRepository{
		db:    db,
		cache: cache.NewCacheRepository[domains.Category](redisClient, "category"),
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

	if err = r.cache.DeleteAll(ctx); err != nil {
		return created, err
	}
	if err = r.cache.Set(ctx, created.ID, created); err != nil {
		return created, err
	}
	return created, nil
}

func (r *categoryRepository) GetByID(ctx context.Context, id int) (*domains.Category, error) {
	category, cacheErr := r.cache.GetByID(ctx, id)
	if cacheErr == nil {
		return category, nil
	}

	query := `
        SELECT id, name, description 
        FROM categories 
        WHERE id = $1`

	category = &domains.Category{}
	row := r.db.QueryRow(ctx, query, id)
	err := row.Scan(&category.ID, &category.Name, &category.Description)
	if err != nil {
		return nil, err
	}

	if cacheErr != redis.Nil {
		return category, err
	}
	if err = r.cache.Set(ctx, category.ID, category); err != nil {
		return category, err
	}
	return category, nil
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

	if err = r.cache.DeleteAll(ctx); err != nil {
		return updated, err
	}
	if err = r.cache.Set(ctx, updated.ID, updated); err != nil {
		return updated, err
	}
	return updated, nil
}

func (r *categoryRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM categories WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
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

func (r *categoryRepository) GetAll(ctx context.Context) ([]*domains.Category, error) {
	categories, cacheErr := r.cache.GetAll(ctx)
	if cacheErr == nil {
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

	var categoriesList []*domains.Category
	for rows.Next() {
		category := &domains.Category{}
		if err = rows.Scan(&category.ID, &category.Name, &category.Description); err != nil {
			return nil, err
		}
		categoriesList = append(categoriesList, category)
	}

	if cacheErr != redis.Nil {
		return categoriesList, err
	}
	if err = r.cache.SetAll(ctx, categoriesList); err != nil {
		return categoriesList, err
	}
	return categoriesList, nil
}
