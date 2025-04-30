package product

import (
	"context"
	"e-commerce/internal/cache"
	"e-commerce/internal/domains"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type ProductRepository interface {
	Create(ctx context.Context, product *domains.Product) (*domains.Product, error)
	GetByID(ctx context.Context, id int) (*domains.Product, error)
	Update(ctx context.Context, id int, product *domains.Product) (*domains.Product, error)
	Delete(ctx context.Context, id int) error
	GetAll(ctx context.Context) ([]*domains.Product, error)
}

type productRepository struct {
	db    *pgxpool.Pool
	cache cache.CachedRepositoryInterface[domains.Product]
	ttl   time.Duration
}

func NewProductRepository(db *pgxpool.Pool, redisClient *redis.Client, ttl time.Duration) ProductRepository {
	return &productRepository{
		db:    db,
		cache: cache.NewBaseCachedRepository[domains.Product](redisClient, "product"),
		ttl:   ttl,
	}
}

func (r *productRepository) Create(ctx context.Context, product *domains.Product) (*domains.Product, error) {
	insertQuery := `
        INSERT INTO products (name, description, price, category_id, brand_id, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, NOW(), NOW()) RETURNING id`

	var id int
	err := r.db.QueryRow(ctx, insertQuery,
		product.Name, product.Description, product.Price,
		product.CategoryID, product.BrandID).Scan(&id)
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

func (r *productRepository) GetByID(ctx context.Context, id int) (*domains.Product, error) {
	if product, err := r.cache.GetByID(ctx, id); err == nil {
		return product, nil
	}

	query := `
        SELECT 
            p.id, p.name, p.description, p.price, p.category_id, p.brand_id, 
            c.name AS category_name,
            b.name AS brand_name,
            p.created_at, p.updated_at
        FROM products p
        LEFT JOIN categories c ON p.category_id = c.id
        LEFT JOIN brands b ON p.brand_id = b.id
        WHERE p.id = $1`

	row := r.db.QueryRow(ctx, query, id)
	product, err := scanProductRow(row)
	if err != nil {
		return nil, err
	}

	_ = r.cache.Set(ctx, product.ID, product, r.ttl)
	return product, nil
}

func scanProductRow(row pgx.Row) (*domains.Product, error) {
	product := &domains.Product{}
	var categoryName, brandName *string

	err := row.Scan(
		&product.ID, &product.Name, &product.Description, &product.Price,
		&product.CategoryID, &product.BrandID,
		&categoryName, &brandName,
		&product.CreatedAt, &product.UpdatedAt,
	)

	if categoryName != nil {
		product.Category = &domains.Category{Name: *categoryName}
	}
	if brandName != nil {
		product.Brand = &domains.Brand{Name: *brandName}
	}

	return product, err
}

func (r *productRepository) Update(ctx context.Context, id int, product *domains.Product) (*domains.Product, error) {
	updateQuery := `
        UPDATE products SET 
            name = $1, description = $2, price = $3, 
            category_id = $4, brand_id = $5, updated_at = NOW() 
        WHERE id = $6`

	_, err := r.db.Exec(ctx, updateQuery,
		product.Name, product.Description, product.Price,
		product.CategoryID, product.BrandID, id)
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

func (r *productRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM products WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	_ = r.cache.Delete(ctx, id)
	_ = r.cache.DeleteAll(ctx)
	return nil
}

func (r *productRepository) GetAll(ctx context.Context) ([]*domains.Product, error) {
	if products, err := r.cache.GetAll(ctx); err == nil {
		return products, nil
	}

	query := `
        SELECT id, name, price, category_id, brand_id, created_at, updated_at 
        FROM products`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*domains.Product
	for rows.Next() {
		product := &domains.Product{}
		err := rows.Scan(
			&product.ID, &product.Name, &product.Price,
			&product.CategoryID, &product.BrandID,
			&product.CreatedAt, &product.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	_ = r.cache.SetAll(ctx, products, r.ttl)
	return products, nil
}
