package product

import (
	"context"
	"e-commerce/internal/cache"
	"e-commerce/internal/domains"

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
	cache cache.CacheRepository[domains.Product]
}

func NewProductRepository(db *pgxpool.Pool, redisClient *redis.Client) ProductRepository {
	return &productRepository{
		db:    db,
		cache: cache.NewCacheRepository[domains.Product](redisClient, "product"),
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

	if err = r.cache.DeleteAll(ctx); err != nil {
		return created, err
	}
	if err = r.cache.Set(ctx, created.ID, created); err != nil {
		return created, err
	}
	return created, nil
}

func (r *productRepository) GetByID(ctx context.Context, id int) (*domains.Product, error) {
	product, cacheErr := r.cache.GetByID(ctx, id)
	if cacheErr == nil {
		return product, nil
	}

	query := `
		SELECT 
		p.id, p.name, p.description, p.price, p.category_id, c.name AS category_name,
		p.brand_id, b.name AS brand_name, p.created_at, p.updated_at,
		(SELECT ARRAY_AGG(s.id ORDER BY s.id) FROM product_skin_types pst 
			JOIN skin_types s ON s.id = pst.skin_type_id WHERE pst.product_id = p.id) AS skin_type_ids,
		(SELECT ARRAY_AGG(s.name ORDER BY s.id) FROM product_skin_types pst 
			JOIN skin_types s ON s.id = pst.skin_type_id WHERE pst.product_id = p.id) AS skin_type_names
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		LEFT JOIN brands b ON p.brand_id = b.id
		WHERE p.id = $1`

	row := r.db.QueryRow(ctx, query, id)
	product, err := scanProductRow(row)
	if err != nil {
		return nil, err
	}

	if cacheErr != redis.Nil {
		return product, err
	}
	if err = r.cache.Set(ctx, product.ID, product); err != nil {
		return product, err
	}
	return product, nil
}

func scanProductRow(row pgx.Row) (*domains.Product, error) {
	product := &domains.Product{}
	var categoryName, brandName *string
	var skinTypeIDs []int
	var skinTypeNames []string

	err := row.Scan(
		&product.ID, &product.Name, &product.Description, &product.Price,
		&product.CategoryID, &categoryName, &product.BrandID, &brandName,
		&product.CreatedAt, &product.UpdatedAt, &skinTypeIDs, &skinTypeNames)
	if err != nil {
		return nil, err
	}

	for i, name := range skinTypeNames {
		product.SkinTypes = append(product.SkinTypes, domains.SkinType{
			ID:   skinTypeIDs[i],
			Name: name,
		})
	}

	if categoryName != nil {
		product.Category = &domains.Category{ID: *product.CategoryID, Name: *categoryName}
	}
	if brandName != nil {
		product.Brand = &domains.Brand{ID: *product.BrandID, Name: *brandName}
	}
	return product, nil
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

	if err = r.cache.DeleteAll(ctx); err != nil {
		return updated, err
	}
	if err = r.cache.Set(ctx, updated.ID, updated); err != nil {
		return updated, err
	}
	return updated, nil
}

func (r *productRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM products WHERE id = $1`
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

func (r *productRepository) GetAll(ctx context.Context) ([]*domains.Product, error) {
	products, cacheErr := r.cache.GetAll(ctx)
	if cacheErr == nil {
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

	var productsList []*domains.Product
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
		productsList = append(productsList, product)
	}

	if cacheErr != redis.Nil {
		return productsList, err
	}
	if err = r.cache.SetAll(ctx, productsList); err != nil {
		return nil, err
	}
	return productsList, nil
}
