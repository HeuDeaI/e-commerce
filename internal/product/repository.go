package product

import (
	"context"
	"database/sql"
	"e-commerce/internal/cache"
	"e-commerce/internal/domains"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type ProductRepository interface {
	Create(ctx context.Context, product *domains.Product) (*domains.Product, error)
	GetByID(ctx context.Context, id int) (*domains.Product, error)
	Update(ctx context.Context, id int, product *domains.Product) (*domains.Product, error)
	Delete(ctx context.Context, id int) error
	GetAll(ctx context.Context) ([]*domains.Product, error)
	GetByFilter(ctx context.Context, skinTypeIDs []int, brandIDs []int, categoryIDs []int) ([]*domains.Product, error)
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
	tx, err := r.db.Begin(ctx)
	if err != nil {
		logrus.WithError(err).Error("Failed to begin transaction")
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	const insertProductQuery = `
        INSERT INTO products (name, description, price, category_id, brand_id)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, name, description, price, category_id, brand_id, created_at, updated_at`

	createdProduct := &domains.Product{}
	err = tx.QueryRow(ctx, insertProductQuery,
		product.Name,
		product.Description,
		product.Price,
		product.CategoryID,
		product.BrandID,
	).Scan(
		&createdProduct.ID,
		&createdProduct.Name,
		&createdProduct.Description,
		&createdProduct.Price,
		&createdProduct.CategoryID,
		&createdProduct.BrandID,
		&createdProduct.CreatedAt,
		&createdProduct.UpdatedAt,
	)
	if err != nil {
		logrus.WithError(err).WithField("product", product).Error("Failed to insert product")
		return nil, err
	}

	if len(product.SkinTypeIDs) > 0 {
		const insertSkinTypeQuery = `
            INSERT INTO product_skin_types (product_id, skin_type_id)
            VALUES ($1, $2)`
		for _, skinTypeID := range product.SkinTypeIDs {
			if _, err = tx.Exec(ctx, insertSkinTypeQuery, createdProduct.ID, skinTypeID); err != nil {
				logrus.WithError(err).Errorf("Failed to insert product_skin_type (product_id: %d, skin_type_id: %d)", createdProduct.ID, skinTypeID)
				return nil, err
			}
		}
		createdProduct.SkinTypeIDs = product.SkinTypeIDs
	}

	if err = tx.Commit(ctx); err != nil {
		logrus.WithError(err).Error("Failed to commit transaction")
		return nil, err
	}

	go func(p *domains.Product) {
		if err := r.cache.SetByID(context.Background(), p.ID, p); err != nil {
			logrus.Warnf("Failed to cache created product asynchronously (ID: %d): %v", p.ID, err)
		} else {
			logrus.Debugf("Successfully cached created product asynchronously (ID: %d)", p.ID)
		}
	}(createdProduct)

	logrus.Debugf("Product created successfully (ID: %d)", createdProduct.ID)
	return createdProduct, nil
}

func (r *productRepository) GetByID(ctx context.Context, id int) (*domains.Product, error) {
	product, err := r.cache.GetByID(ctx, id)
	if err == nil {
		logrus.Debugf("Cache hit for product (ID: %d)", id)
		return product, nil
	}
	if !errors.Is(err, redis.Nil) {
		logrus.Errorf("Cache lookup failed for product (ID: %d): %v", id, err)
	}

	const getQuery = `
        SELECT 
            p.id, p.name, p.description, p.price, 
            COALESCE(p.category_id, c.id) AS category_id, c.name AS category_name,
            COALESCE(p.brand_id, b.id) AS brand_id, b.name AS brand_name, 
            p.created_at, p.updated_at,
            COALESCE(ARRAY_AGG(st.id ORDER BY st.id), '{}') AS skin_type_ids,
            COALESCE(ARRAY_AGG(st.name ORDER BY st.id), '{}') AS skin_type_names
        FROM products p
        LEFT JOIN categories c ON p.category_id = c.id
        LEFT JOIN brands b ON p.brand_id = b.id
        LEFT JOIN product_skin_types pst ON p.id = pst.product_id
        LEFT JOIN skin_types st ON pst.skin_type_id = st.id
        WHERE p.id = $1
        GROUP BY p.id, c.id, b.id`

	rows, err := r.db.Query(ctx, getQuery, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logrus.Infof("Product not found (ID: %d)", id)
			return nil, sql.ErrNoRows
		}
		logrus.Errorf("Failed to get product (ID: %d): %v", id, err)
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		logrus.Infof("Product not found (ID: %d)", id)
		return nil, sql.ErrNoRows
	}

	product = &domains.Product{
		Category: &domains.Category{},
		Brand:    &domains.Brand{},
	}
	var skinTypeIDs []int
	var skinTypeNames []string

	err = rows.Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.Category.ID,
		&product.Category.Name,
		&product.Brand.ID,
		&product.Brand.Name,
		&product.CreatedAt,
		&product.UpdatedAt,
		&skinTypeIDs,
		&skinTypeNames,
	)
	if err != nil {
		logrus.Errorf("Failed to scan product row (ID: %d): %v", id, err)
		return nil, err
	}

	if product.Category != nil {
		product.CategoryID = &product.Category.ID
	} else {
		product.CategoryID = nil
	}

	if product.Brand != nil {
		product.BrandID = &product.Brand.ID
	} else {
		product.BrandID = nil
	}

	for i, id := range skinTypeIDs {
		product.SkinTypes = append(product.SkinTypes, domains.SkinType{
			ID:   id,
			Name: skinTypeNames[i],
		})
	}

	logrus.Debugf("Product retrieved successfully (ID: %d)", product.ID)

	go func(p *domains.Product) {
		if err := r.cache.SetByID(context.Background(), p.ID, p); err != nil {
			logrus.Warnf("Failed to cache product asynchronously (ID: %d): %v", p.ID, err)
		} else {
			logrus.Debugf("Successfully cached product asynchronously (ID: %d)", p.ID)
		}
	}(product)

	return product, nil
}

func (r *productRepository) Update(ctx context.Context, id int, product *domains.Product) (*domains.Product, error) {
	const updateQuery = `
        UPDATE products 
        SET name = $1, description = $2, price = $3, 
            category_id = $4, brand_id = $5 
        WHERE id = $6
        RETURNING id, name, description, price, category_id, brand_id, created_at, updated_at`

	updatedProduct := &domains.Product{}
	err := r.db.QueryRow(ctx, updateQuery,
		product.Name,
		product.Description,
		product.Price,
		product.CategoryID,
		product.BrandID,
		id,
	).Scan(
		&updatedProduct.ID,
		&updatedProduct.Name,
		&updatedProduct.Description,
		&updatedProduct.Price,
		&updatedProduct.CategoryID,
		&updatedProduct.BrandID,
		&updatedProduct.CreatedAt,
		&updatedProduct.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logrus.Infof("Attempted to update non-existent product (ID: %d)", id)
			return nil, sql.ErrNoRows
		}
		logrus.Errorf("Failed to update product (ID: %d): %v", id, err)
		return nil, err
	}

	logrus.Debugf("Product updated successfully (ID: %d)", updatedProduct.ID)

	if err := r.cache.DeleteAll(ctx); err != nil {
		logrus.Warnf("Failed to clear product cache after update (ID: %d): %v", id, err)
	}
	go func(p *domains.Product) {
		if err := r.cache.SetByID(context.Background(), p.ID, p); err != nil {
			logrus.Warnf("Failed to cache updated product asynchronously (ID: %d): %v", p.ID, err)
		} else {
			logrus.Debugf("Successfully cached updated product asynchronously (ID: %d)", p.ID)
		}
	}(updatedProduct)

	return updatedProduct, nil
}

func (r *productRepository) Delete(ctx context.Context, id int) error {
	const deleteQuery = `DELETE FROM products WHERE id = $1 RETURNING id`

	var deletedID int
	err := r.db.QueryRow(ctx, deleteQuery, id).Scan(&deletedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logrus.Infof("Attempted to delete non-existent product (ID: %d)", id)
			return sql.ErrNoRows
		}
		logrus.Errorf("Failed to delete product (ID: %d): %v", id, err)
		return err
	}

	logrus.Debugf("Product deleted successfully (ID: %d)", deletedID)

	if err := r.cache.Delete(ctx, id); err != nil {
		logrus.Warnf("Failed to remove product from cache (ID: %d): %v", id, err)
	}
	if err := r.cache.DeleteAll(ctx); err != nil {
		logrus.Warnf("Failed to clear all products cache after deletion (ID: %d): %v", id, err)
	}

	return nil
}

func (r *productRepository) GetAll(ctx context.Context) ([]*domains.Product, error) {
	products, err := r.cache.GetAll(ctx)
	if err == nil {
		logrus.Debug("Cache hit for all products")
		return products, nil
	}
	if !errors.Is(err, redis.Nil) {
		logrus.Errorf("Cache lookup failed for all products: %v", err)
	}

	const getAllQuery = `
        SELECT id, name, price
        FROM products`

	rows, err := r.db.Query(ctx, getAllQuery)
	if err != nil {
		logrus.Errorf("Failed to get all products: %v", err)
		return nil, err
	}
	defer rows.Close()

	var productsList []*domains.Product
	for rows.Next() {
		product := &domains.Product{}
		err := rows.Scan(
			&product.ID, &product.Name, &product.Price,
		)
		if err != nil {
			logrus.Errorf("Failed to scan product row: %v", err)
			return nil, err
		}
		productsList = append(productsList, product)
	}
	if rows.Err() != nil {
		logrus.Errorf("Error occurred during iteration of rows: %v", rows.Err())
		return nil, rows.Err()
	}

	logrus.Debugf("All products retrieved successfully (Count: %d)", len(productsList))

	go func(pl []*domains.Product) {
		if err := r.cache.SetAll(context.Background(), pl); err != nil {
			logrus.Warnf("Failed to cache all products asynchronously: %v", err)
		} else {
			logrus.Debugf("Successfully cached all products asynchronously (Count: %d)", len(pl))
		}
	}(productsList)

	return productsList, nil
}

func (r *productRepository) GetByFilter(ctx context.Context, skinTypeIDs []int, brandIDs []int, categoryIDs []int,
) ([]*domains.Product, error) {
	filterKey := fmt.Sprintf("filter:skin=%v:brand=%v:category=%v", skinTypeIDs, brandIDs, categoryIDs)

	products, err := r.cache.GetByKey(ctx, filterKey)
	if err == nil {
		logrus.Debugf("Cache hit for products by filter (key: %s)", filterKey)
		return products, nil
	} else {
		if !errors.Is(err, redis.Nil) {
			logrus.Errorf("Cache lookup failed for filter key %s: %v", filterKey, err)
		} else {
			logrus.Debugf("Cache miss for products by filter (key: %s)", filterKey)
		}
	}

	var (
		queryBuilder strings.Builder
		args         []interface{}
		conditions   []string
	)

	queryBuilder.WriteString("SELECT p.id, p.name, p.description, p.price FROM products p")
	if len(skinTypeIDs) > 0 {
		queryBuilder.WriteString(" JOIN product_skin_types pst ON p.id = pst.product_id")
	}

	argPos := 1
	if len(brandIDs) > 0 {
		conditions = append(conditions, fmt.Sprintf("p.brand_id = ANY($%d)", argPos))
		args = append(args, brandIDs)
		argPos++
	}

	if len(categoryIDs) > 0 {
		conditions = append(conditions, fmt.Sprintf("p.category_id = ANY($%d)", argPos))
		args = append(args, categoryIDs)
		argPos++
	}

	if len(skinTypeIDs) > 0 {
		conditions = append(conditions, fmt.Sprintf("pst.skin_type_id = ANY($%d)", argPos))
		args = append(args, skinTypeIDs)
		argPos++
	}

	if len(conditions) > 0 {
		queryBuilder.WriteString(" WHERE " + strings.Join(conditions, " AND "))
	}

	query := queryBuilder.String()
	logrus.Debugf("Filter Query: %s, Args: %+v", query, args)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		logrus.Errorf("Failed to execute filtered query: %v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p domains.Product
		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Description,
			&p.Price,
		); err != nil {
			logrus.Errorf("Failed to scan product row in filter query: %v", err)
			return nil, err
		}
		products = append(products, &p)
	}
	if err := rows.Err(); err != nil {
		logrus.Errorf("Error iterating filter query result rows: %v", err)
		return nil, err
	}

	logrus.Debugf("Filter products retrieved successfully (Count: %d)", len(products))

	go func(prodList []*domains.Product, key string) {
		if err := r.cache.SetByKey(context.Background(), key, prodList); err != nil {
			logrus.Warnf("Failed to cache filtered products asynchronously (key: %s): %v", key, err)
		} else {
			logrus.Debugf("Successfully cached filtered products asynchronously (key: %s)", key)
		}
	}(products, filterKey)

	return products, nil
}
