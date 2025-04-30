package brand

import (
	"context"
	"e-commerce/internal/domains"

	"github.com/jackc/pgx/v5/pgxpool"
)

type BrandRepository interface {
	CreateBrand(ctx context.Context, brand *domains.Brand) (*domains.Brand, error)
	GetBrandByID(ctx context.Context, id int) (*domains.Brand, error)
	UpdateBrand(ctx context.Context, id int, brand *domains.Brand) (*domains.Brand, error)
	DeleteBrand(ctx context.Context, id int) error
	GetAllBrands(ctx context.Context) ([]*domains.Brand, error)
}

type brandRepository struct {
	pool *pgxpool.Pool
}

func NewBrandRepository(pool *pgxpool.Pool) BrandRepository {
	return &brandRepository{pool: pool}
}

func (r *brandRepository) CreateBrand(ctx context.Context, brand *domains.Brand) (*domains.Brand, error) {
	query := `
        INSERT INTO brands (name, description, website)
        VALUES ($1, $2, $3) RETURNING id`

	row := r.pool.QueryRow(ctx, query, brand.Name, brand.Description, brand.Website)
	err := row.Scan(&brand.ID)
	if err != nil {
		return nil, err
	}

	return brand, nil
}

func (r *brandRepository) GetBrandByID(ctx context.Context, id int) (*domains.Brand, error) {
	query := `SELECT id, name, description, website FROM brands WHERE id = $1`

	brand := &domains.Brand{}
	row := r.pool.QueryRow(ctx, query, id)
	err := row.Scan(&brand.ID, &brand.Name, &brand.Description, &brand.Website)
	if err != nil {
		return nil, err
	}

	return brand, nil
}

func (r *brandRepository) UpdateBrand(ctx context.Context, id int, brand *domains.Brand) (*domains.Brand, error) {
	query := `
        UPDATE brands SET name = $1, description = $2, website = $3 
        WHERE id = $4 RETURNING id`

	row := r.pool.QueryRow(ctx, query, brand.Name, brand.Description, brand.Website, id)
	err := row.Scan(&brand.ID)
	if err != nil {
		return nil, err
	}

	return brand, nil
}

func (r *brandRepository) DeleteBrand(ctx context.Context, id int) error {
	query := `DELETE FROM brands WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	return nil
}

func (r *brandRepository) GetAllBrands(ctx context.Context) ([]*domains.Brand, error) {
	query := `SELECT id, name, description, website FROM brands`

	rows, err := r.pool.Query(ctx, query)
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

	return brands, nil
}
