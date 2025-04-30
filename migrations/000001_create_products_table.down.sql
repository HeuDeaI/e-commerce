DROP INDEX IF EXISTS idx_product_images_is_main;
DROP INDEX IF EXISTS idx_product_images_unique_main;
DROP INDEX IF EXISTS idx_product_images_prod_id;

DROP INDEX IF EXISTS idx_products_brand_id;
DROP INDEX IF EXISTS idx_products_category_id;

DROP TABLE IF EXISTS product_images;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS brands;
DROP TABLE IF EXISTS categories;
