package category

import (
	"context"
	"database/sql"
	"e-commerce/internal/cache"
	"e-commerce/internal/domains"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
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
	const insertQuery = `
        INSERT INTO categories (name, description)
        VALUES ($1, $2)
        RETURNING id, name, description`

	createdCategory := &domains.Category{}
	err := r.db.QueryRow(ctx, insertQuery, category.Name, category.Description).Scan(
		&createdCategory.ID,
		&createdCategory.Name,
		&createdCategory.Description,
	)
	if err != nil {
		logrus.WithError(err).WithField("category", category).Error("Failed to insert category")
		return nil, err
	}

	logrus.Debugf("Category created successfully (ID: %d)", createdCategory.ID)

	if err := r.cache.DeleteAll(ctx); err != nil {
		logrus.Warnf("Failed to clear category cache after creation (ID: %d): %v", createdCategory.ID, err)
	}
	go func(c *domains.Category) {
		if err := r.cache.SetByID(context.Background(), c.ID, c); err != nil {
			logrus.Warnf("Failed to cache created category asynchronously (ID: %d): %v", c.ID, err)
		} else {
			logrus.Debugf("Successfully cached created category asynchronously (ID: %d)", c.ID)
		}
	}(createdCategory)

	return createdCategory, nil
}

func (r *categoryRepository) GetByID(ctx context.Context, id int) (*domains.Category, error) {
	category, err := r.cache.GetByID(ctx, id)
	if err == nil {
		logrus.Debugf("Cache hit for category (ID: %d)", id)
		return category, nil
	}
	if !errors.Is(err, redis.Nil) {
		logrus.Errorf("Cache lookup failed for category (ID: %d): %v", id, err)
	}

	const getQuery = `SELECT id, name, description FROM categories WHERE id = $1`
	category = &domains.Category{}
	err = r.db.QueryRow(ctx, getQuery, id).Scan(
		&category.ID,
		&category.Name,
		&category.Description,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logrus.Infof("Category not found (ID: %d)", id)
			return nil, sql.ErrNoRows
		}
		logrus.Errorf("Failed to get category (ID: %d): %v", id, err)
		return nil, err
	}

	logrus.Debugf("Category retrieved successfully (ID: %d)", category.ID)

	go func(c *domains.Category) {
		if err := r.cache.SetByID(context.Background(), c.ID, c); err != nil {
			logrus.Warnf("Failed to cache category asynchronously (ID: %d): %v", c.ID, err)
		} else {
			logrus.Debugf("Successfully cached category asynchronously (ID: %d)", c.ID)
		}
	}(category)

	return category, nil
}

func (r *categoryRepository) Update(ctx context.Context, id int, category *domains.Category) (*domains.Category, error) {
	const updateQuery = `
        UPDATE categories 
        SET name = $1, description = $2 
        WHERE id = $3
        RETURNING id, name, description`

	updatedCategory := &domains.Category{}
	err := r.db.QueryRow(ctx, updateQuery,
		category.Name,
		category.Description,
		id,
	).Scan(
		&updatedCategory.ID,
		&updatedCategory.Name,
		&updatedCategory.Description,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logrus.Infof("Attempted to update non-existent category (ID: %d)", id)
			return nil, sql.ErrNoRows
		}
		logrus.Errorf("Failed to update category (ID: %d): %v", id, err)
		return nil, err
	}

	logrus.Debugf("Category updated successfully (ID: %d)", updatedCategory.ID)

	if err := r.cache.DeleteAll(ctx); err != nil {
		logrus.Warnf("Failed to clear category cache after update (ID: %d): %v", id, err)
	}
	go func(c *domains.Category) {
		if err := r.cache.SetByID(context.Background(), c.ID, c); err != nil {
			logrus.Warnf("Failed to cache updated category asynchronously (ID: %d): %v", c.ID, err)
		} else {
			logrus.Debugf("Successfully cached updated category asynchronously (ID: %d)", c.ID)
		}
	}(updatedCategory)

	return updatedCategory, nil
}

func (r *categoryRepository) Delete(ctx context.Context, id int) error {
	const deleteQuery = `DELETE FROM categories WHERE id = $1 RETURNING id`

	var deletedID int
	err := r.db.QueryRow(ctx, deleteQuery, id).Scan(&deletedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logrus.Infof("Attempted to delete non-existent category (ID: %d)", id)
			return sql.ErrNoRows
		}
		logrus.Errorf("Failed to delete category (ID: %d): %v", id, err)
		return err
	}

	logrus.Debugf("Category deleted successfully (ID: %d)", deletedID)

	if err := r.cache.Delete(ctx, id); err != nil {
		logrus.Warnf("Failed to remove category from cache (ID: %d): %v", id, err)
	}
	if err := r.cache.DeleteAll(ctx); err != nil {
		logrus.Warnf("Failed to clear all categories cache after deletion (ID: %d): %v", id, err)
	}

	return nil
}

func (r *categoryRepository) GetAll(ctx context.Context) ([]*domains.Category, error) {
	categories, err := r.cache.GetAll(ctx)
	if err == nil {
		logrus.Debug("Cache hit for all categories")
		return categories, nil
	}
	if !errors.Is(err, redis.Nil) {
		logrus.Errorf("Cache lookup failed for all categories: %v", err)
	}

	const getAllQuery = `SELECT id, name, description FROM categories`
	rows, err := r.db.Query(ctx, getAllQuery)
	if err != nil {
		logrus.Errorf("Failed to get all categories: %v", err)
		return nil, err
	}
	defer rows.Close()

	var categoriesList []*domains.Category
	for rows.Next() {
		category := &domains.Category{}
		if err = rows.Scan(
			&category.ID,
			&category.Name,
			&category.Description,
		); err != nil {
			logrus.Errorf("Failed to scan category record: %v", err)
			return nil, err
		}
		categoriesList = append(categoriesList, category)
	}
	if rows.Err() != nil {
		logrus.Errorf("Error occurred during iteration of rows: %v", rows.Err())
		return nil, rows.Err()
	}

	logrus.Debugf("All categories retrieved successfully (Count: %d)", len(categoriesList))

	go func(cl []*domains.Category) {
		if err := r.cache.SetAll(context.Background(), cl); err != nil {
			logrus.Warnf("Failed to cache all categories asynchronously: %v", err)
		} else {
			logrus.Debugf("Successfully cached all categories asynchronously (Count: %d)", len(cl))
		}
	}(categoriesList)

	return categoriesList, nil
}
