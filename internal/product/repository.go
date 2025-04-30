package product

import (
	"context"
	"e-commerce/internal/domains"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductRepository interface {
	CreateProduct(ctx context.Context, product *domains.Product) (*domains.Product, error)
	GetProductByID(ctx context.Context, id int) (*domains.Product, error)
	UpdateProduct(ctx context.Context, id int, product *domains.Product) (*domains.Product, error)
	DeleteProduct(ctx context.Context, id int) error
	GetAllProducts(ctx context.Context) ([]*domains.Product, error)
}

type productRepository struct {
	pool *pgxpool.Pool
}

func NewProductRepository(pool *pgxpool.Pool) ProductRepository {
	return &productRepository{pool: pool}
}

func (r *productRepository) CreateProduct(ctx context.Context, product *domains.Product) (*domains.Product, error) {
	query := `
        INSERT INTO products (name, description, price, category_id, brand_id, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, NOW(), NOW()) RETURNING id, created_at, updated_at`

	row := r.pool.QueryRow(ctx, query, product.Name, product.Description, product.Price, product.CategoryID, product.BrandID)
	err := row.Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return product, nil
}

func (r *productRepository) GetProductByID(ctx context.Context, id int) (*domains.Product, error) {
	query := `
        SELECT id, name, description, price, category_id, brand_id, created_at, updated_at 
        FROM products WHERE id = $1`

	product := &domains.Product{}
	row := r.pool.QueryRow(ctx, query, id)
	err := row.Scan(
		&product.ID, &product.Name, &product.Description, &product.Price,
		&product.CategoryID, &product.BrandID, &product.CreatedAt, &product.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return product, nil
}

func (r *productRepository) UpdateProduct(ctx context.Context, id int, product *domains.Product) (*domains.Product, error) {
	query := `
        UPDATE products SET name = $1, description = $2, price = $3, category_id = $4, brand_id = $5, updated_at = NOW() 
        WHERE id = $6 RETURNING updated_at`

	row := r.pool.QueryRow(ctx, query, product.Name, product.Description, product.Price, product.CategoryID, product.BrandID, id)
	err := row.Scan(&product.UpdatedAt)
	if err != nil {
		return nil, err
	}

	product.ID = id
	return product, nil
}

func (r *productRepository) DeleteProduct(ctx context.Context, id int) error {
	query := `DELETE FROM products WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	return nil
}

func (r *productRepository) GetAllProducts(ctx context.Context) ([]*domains.Product, error) {
	query := `
        SELECT id, name, description, price, category_id, brand_id, created_at, updated_at 
        FROM products`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*domains.Product
	for rows.Next() {
		product := &domains.Product{}
		err := rows.Scan(
			&product.ID, &product.Name, &product.Description, &product.Price,
			&product.CategoryID, &product.BrandID, &product.CreatedAt, &product.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	return products, nil
}
