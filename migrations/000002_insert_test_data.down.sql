-- Begin transaction for rollback
BEGIN;

-- Delete from join table first (child table)
DELETE FROM product_skin_types;

-- Delete dependent rows from product_images
DELETE FROM product_images;

-- Delete products which are referenced by images and join table
DELETE FROM products;

-- Delete from independent reference tables.
DELETE FROM skin_types;
DELETE FROM brands;
DELETE FROM categories;

-- Commit transaction to save changes
COMMIT;
