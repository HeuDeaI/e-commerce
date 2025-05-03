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
	Create(ctx context.Context, req *domains.ProductRequest) (*domains.ProductResponse, error)
	GetByID(ctx context.Context, id int) (*domains.ProductResponse, error)
	Update(ctx context.Context, id int, req *domains.ProductRequest) (*domains.ProductResponse, error)
	Delete(ctx context.Context, id int) error
	GetAll(ctx context.Context) ([]*domains.ProductResponse, error)
	GetByFilter(ctx context.Context, skinTypeIDs []int, brandIDs []int, categoryIDs []int) ([]*domains.ProductResponse, error)
}

type productRepository struct {
	db    *pgxpool.Pool
	cache cache.CacheRepository[domains.ProductResponse]
}

func NewProductRepository(db *pgxpool.Pool, redisClient *redis.Client) ProductRepository {
	return &productRepository{
		db:    db,
		cache: cache.NewCacheRepository[domains.ProductResponse](redisClient, "product"),
	}
}

func (r *productRepository) Create(ctx context.Context, req *domains.ProductRequest) (*domains.ProductResponse, error) {
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

	var prodResp domains.ProductResponse
	var tempCategoryID, tempBrandID sql.NullInt64

	err = tx.QueryRow(ctx, insertProductQuery,
		req.Name,
		req.Description,
		req.Price,
		req.CategoryID,
		req.BrandID,
	).Scan(
		&prodResp.ID,
		&prodResp.Name,
		&prodResp.Description,
		&prodResp.Price,
		&tempCategoryID,
		&tempBrandID,
		&prodResp.CreatedAt,
		&prodResp.UpdatedAt,
	)
	if err != nil {
		logrus.WithError(err).WithField("req", req).Error("Failed to insert product")
		return nil, err
	}

	if tempCategoryID.Valid {
		prodResp.Category = &domains.Category{ID: int(tempCategoryID.Int64)}
	}
	if tempBrandID.Valid {
		prodResp.Brand = &domains.Brand{ID: int(tempBrandID.Int64)}
	}

	if len(req.SkinTypeIDs) > 0 {
		const insertSkinTypeQuery = `
            INSERT INTO product_skin_types (product_id, skin_type_id)
            VALUES ($1, $2)`
		for _, skinTypeID := range req.SkinTypeIDs {
			if _, err = tx.Exec(ctx, insertSkinTypeQuery, prodResp.ID, skinTypeID); err != nil {
				logrus.WithError(err).Errorf("Failed to insert product_skin_type (product_id: %d, skin_type_id: %d)", prodResp.ID, skinTypeID)
				return nil, err
			}
			prodResp.SkinTypes = append(prodResp.SkinTypes, domains.SkinType{ID: skinTypeID})
		}
	}

	if err = tx.Commit(ctx); err != nil {
		logrus.WithError(err).Error("Failed to commit transaction")
		return nil, err
	}

	go func(p *domains.ProductResponse) {
		if err := r.cache.SetByID(context.Background(), p.ID, p); err != nil {
			logrus.Warnf("Failed to cache created product asynchronously (ID: %d): %v", p.ID, err)
		} else {
			logrus.Debugf("Successfully cached created product asynchronously (ID: %d)", p.ID)
		}
	}(&prodResp)

	logrus.Debugf("Product created successfully (ID: %d)", prodResp.ID)
	return &prodResp, nil
}

func (r *productRepository) GetByID(ctx context.Context, id int) (*domains.ProductResponse, error) {
	prodResp, err := r.cache.GetByID(ctx, id)
	if err == nil {
		logrus.Debugf("Cache hit for product (ID: %d)", id)
		return prodResp, nil
	}
	if !errors.Is(err, redis.Nil) {
		logrus.Errorf("Cache lookup failed for product (ID: %d): %v", id, err)
	}

	const getQuery = `
        SELECT 
            p.id, p.name, p.description, p.price, 
            c.id AS c_id, c.name AS c_name,
            b.id AS b_id, b.name AS b_name, 
            p.created_at, p.updated_at,
            COALESCE(ARRAY_AGG(st.id ORDER BY st.id) FILTER (WHERE st.id IS NOT NULL), '{}') AS skin_type_ids,
            COALESCE(ARRAY_AGG(st.name ORDER BY st.id) FILTER (WHERE st.name IS NOT NULL), '{}') AS skin_type_names
        FROM products p
        LEFT JOIN categories c ON p.category_id = c.id
        LEFT JOIN brands b ON p.brand_id = b.id
        LEFT JOIN product_skin_types pst ON p.id = pst.product_id
        LEFT JOIN skin_types st ON pst.skin_type_id = st.id
        WHERE p.id = $1
        GROUP BY p.id, c.id, b.id`

	row := r.db.QueryRow(ctx, getQuery, id)

	prodResp = &domains.ProductResponse{
		Category: &domains.Category{},
		Brand:    &domains.Brand{},
	}
	var skinTypeIDs []int
	var skinTypeNames []string

	err = row.Scan(
		&prodResp.ID,
		&prodResp.Name,
		&prodResp.Description,
		&prodResp.Price,
		&prodResp.Category.ID,
		&prodResp.Category.Name,
		&prodResp.Brand.ID,
		&prodResp.Brand.Name,
		&prodResp.CreatedAt,
		&prodResp.UpdatedAt,
		&skinTypeIDs,
		&skinTypeNames,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logrus.Infof("Product not found (ID: %d)", id)
			return nil, sql.ErrNoRows
		}
		logrus.Errorf("Failed to scan product row (ID: %d): %v", id, err)
		return nil, err
	}

	for i, stID := range skinTypeIDs {
		prodResp.SkinTypes = append(prodResp.SkinTypes, domains.SkinType{
			ID:   stID,
			Name: skinTypeNames[i],
		})
	}

	go func(p *domains.ProductResponse) {
		if err := r.cache.SetByID(context.Background(), p.ID, p); err != nil {
			logrus.Warnf("Failed to cache product asynchronously (ID: %d): %v", p.ID, err)
		} else {
			logrus.Debugf("Successfully cached product asynchronously (ID: %d)", p.ID)
		}
	}(prodResp)

	logrus.Debugf("Product retrieved successfully (ID: %d)", prodResp.ID)
	return prodResp, nil
}

func (r *productRepository) Update(ctx context.Context, id int, req *domains.ProductRequest) (*domains.ProductResponse, error) {
	const updateQuery = `
        UPDATE products 
        SET name = $1, description = $2, price = $3, 
            category_id = $4, brand_id = $5 
        WHERE id = $6
        RETURNING id, name, description, price, category_id, brand_id, created_at, updated_at`

	var prodResp domains.ProductResponse
	var tempCategoryID, tempBrandID sql.NullInt64

	err := r.db.QueryRow(ctx, updateQuery,
		req.Name,
		req.Description,
		req.Price,
		req.CategoryID,
		req.BrandID,
		id,
	).Scan(
		&prodResp.ID,
		&prodResp.Name,
		&prodResp.Description,
		&prodResp.Price,
		&tempCategoryID,
		&tempBrandID,
		&prodResp.CreatedAt,
		&prodResp.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logrus.Infof("Attempted to update non-existent product (ID: %d)", id)
			return nil, sql.ErrNoRows
		}
		logrus.Errorf("Failed to update product (ID: %d): %v", id, err)
		return nil, err
	}

	if tempCategoryID.Valid {
		prodResp.Category = &domains.Category{ID: int(tempCategoryID.Int64)}
	}
	if tempBrandID.Valid {
		prodResp.Brand = &domains.Brand{ID: int(tempBrandID.Int64)}
	}

	const deleteAssocQuery = `DELETE FROM product_skin_types WHERE product_id = $1`
	if _, err = r.db.Exec(ctx, deleteAssocQuery, id); err != nil {
		logrus.Errorf("Failed to delete old product_skin_types for product (ID: %d): %v", id, err)
		return nil, err
	}
	prodResp.SkinTypes = []domains.SkinType{}

	if len(req.SkinTypeIDs) > 0 {
		const insertSkinTypeQuery = `
            INSERT INTO product_skin_types (product_id, skin_type_id)
            VALUES ($1, $2)`
		for _, skinTypeID := range req.SkinTypeIDs {
			if _, err = r.db.Exec(ctx, insertSkinTypeQuery, id, skinTypeID); err != nil {
				logrus.WithError(err).Errorf("Failed to insert product_skin_type (product_id: %d, skin_type_id: %d)", id, skinTypeID)
				return nil, err
			}
			prodResp.SkinTypes = append(prodResp.SkinTypes, domains.SkinType{ID: skinTypeID})
		}
	}

	if err := r.cache.DeleteAll(ctx); err != nil {
		logrus.Warnf("Failed to clear product cache after update (ID: %d): %v", id, err)
	}
	go func(p *domains.ProductResponse) {
		if err := r.cache.SetByID(context.Background(), p.ID, p); err != nil {
			logrus.Warnf("Failed to cache updated product asynchronously (ID: %d): %v", p.ID, err)
		} else {
			logrus.Debugf("Successfully cached updated product asynchronously (ID: %d)", p.ID)
		}
	}(&prodResp)

	logrus.Debugf("Product updated successfully (ID: %d)", prodResp.ID)
	return &prodResp, nil
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

	if err := r.cache.Delete(ctx, id); err != nil {
		logrus.Warnf("Failed to remove product from cache (ID: %d): %v", id, err)
	}
	if err := r.cache.DeleteAll(ctx); err != nil {
		logrus.Warnf("Failed to clear all products cache after deletion (ID: %d): %v", id, err)
	}

	logrus.Debugf("Product deleted successfully (ID: %d)", deletedID)
	return nil
}

func (r *productRepository) GetAll(ctx context.Context) ([]*domains.ProductResponse, error) {
	productsResp, err := r.cache.GetAll(ctx)
	if err == nil {
		logrus.Debug("Cache hit for all products")
		return productsResp, nil
	}
	if !errors.Is(err, redis.Nil) {
		logrus.Errorf("Cache lookup failed for all products: %v", err)
	}

	const getAllQuery = `
        SELECT id, name, price
        FROM products
        ORDER BY id`

	rows, err := r.db.Query(ctx, getAllQuery)
	if err != nil {
		logrus.Errorf("Failed to query all products: %v", err)
		return nil, err
	}
	defer rows.Close()

	var productsList []*domains.ProductResponse
	for rows.Next() {
		prod := new(domains.ProductResponse)
		if err := rows.Scan(&prod.ID, &prod.Name, &prod.Price); err != nil {
			logrus.Errorf("Failed to scan product row: %v", err)
			return nil, err
		}
		productsList = append(productsList, prod)
	}
	if err := rows.Err(); err != nil {
		logrus.Errorf("Row iteration error: %v", err)
		return nil, err
	}

	go func(pl []*domains.ProductResponse) {
		if err := r.cache.SetAll(context.Background(), pl); err != nil {
			logrus.Warnf("Failed to cache all products asynchronously: %v", err)
		} else {
			logrus.Debugf("Successfully cached all products asynchronously (Count: %d)", len(pl))
		}
	}(productsList)

	logrus.Debugf("All products retrieved successfully (Count: %d)", len(productsList))
	return productsList, nil
}

func (r *productRepository) GetByFilter(ctx context.Context, skinTypeIDs []int, brandIDs []int, categoryIDs []int) ([]*domains.ProductResponse, error) {
	filterKey := fmt.Sprintf("filter:skin=%v:brand=%v:category=%v", skinTypeIDs, brandIDs, categoryIDs)

	productsResp, err := r.cache.GetByKey(ctx, filterKey)
	if err == nil {
		logrus.Debugf("Cache hit for products by filter (key: %s)", filterKey)
		return productsResp, nil
	}
	if !errors.Is(err, redis.Nil) {
		logrus.Errorf("Cache lookup failed for filter key %s: %v", filterKey, err)
	}

	var (
		queryBuilder strings.Builder
		args         []interface{}
		conditions   []string
	)

	queryBuilder.WriteString("SELECT DISTINCT p.id, p.name, p.price FROM products p")
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
	queryBuilder.WriteString(" ORDER BY p.id")

	query := queryBuilder.String()
	logrus.Debugf("Filter Query: %s, Args: %+v", query, args)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		logrus.Errorf("Failed to execute filtered query: %v", err)
		return nil, err
	}
	defer rows.Close()

	var productsList []*domains.ProductResponse
	for rows.Next() {
		prod := new(domains.ProductResponse)
		if err := rows.Scan(&prod.ID, &prod.Name, &prod.Price); err != nil {
			logrus.Errorf("Failed to scan product row in filter query: %v", err)
			return nil, err
		}
		productsList = append(productsList, prod)
	}
	if err = rows.Err(); err != nil {
		logrus.Errorf("Error iterating filter query result rows: %v", err)
		return nil, err
	}

	go func(prodList []*domains.ProductResponse, key string) {
		if err := r.cache.SetByKey(context.Background(), key, prodList); err != nil {
			logrus.Warnf("Failed to cache filtered products asynchronously (key: %s): %v", key, err)
		} else {
			logrus.Debugf("Successfully cached filtered products asynchronously (key: %s)", key)
		}
	}(productsList, filterKey)

	logrus.Debugf("Filter products retrieved successfully (Count: %d)", len(productsList))
	return productsList, nil
}
