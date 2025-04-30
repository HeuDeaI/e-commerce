package category

import (
	"context"

	"e-commerce/internal/domains"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoryRepository interface {
	CreateCategory(ctx context.Context, category *domains.Category) (*domains.Category, error)
	GetCategoryByID(ctx context.Context, id int) (*domains.Category, error)
	UpdateCategory(ctx context.Context, id int, category *domains.Category) (*domains.Category, error)
	DeleteCategory(ctx context.Context, id int) error
	GetAllCategories(ctx context.Context) ([]*domains.Category, error)
}

type categoryRepository struct {
	pool *pgxpool.Pool
}

func NewCategoryRepository(pool *pgxpool.Pool) CategoryRepository {
	return &categoryRepository{pool: pool}
}

func (r *categoryRepository) CreateCategory(ctx context.Context, category *domains.Category) (*domains.Category, error) {
	query := `
        INSERT INTO categories (name, description)
        VALUES ($1, $2)
        RETURNING id`
	row := r.pool.QueryRow(ctx, query, category.Name, category.Description)
	if err := row.Scan(&category.ID); err != nil {
		return nil, err
	}
	return category, nil
}

func (r *categoryRepository) GetCategoryByID(ctx context.Context, id int) (*domains.Category, error) {
	query := `
        SELECT id, name, description
        FROM categories
        WHERE id = $1`
	category := &domains.Category{}
	row := r.pool.QueryRow(ctx, query, id)
	if err := row.Scan(&category.ID, &category.Name, &category.Description); err != nil {
		return nil, err
	}
	return category, nil
}

func (r *categoryRepository) UpdateCategory(ctx context.Context, id int, category *domains.Category) (*domains.Category, error) {
	query := `
        UPDATE categories
        SET name = $1, description = $2
        WHERE id = $3
        RETURNING id`
	row := r.pool.QueryRow(ctx, query, category.Name, category.Description, id)
	if err := row.Scan(&category.ID); err != nil {
		return nil, err
	}
	return category, nil
}

func (r *categoryRepository) DeleteCategory(ctx context.Context, id int) error {
	query := `DELETE FROM categories WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *categoryRepository) GetAllCategories(ctx context.Context) ([]*domains.Category, error) {
	query := `SELECT id, name, description FROM categories`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*domains.Category
	for rows.Next() {
		category := &domains.Category{}
		if err := rows.Scan(&category.ID, &category.Name, &category.Description); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, nil
}
