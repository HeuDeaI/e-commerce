CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    CHECK (name <> '')
);

CREATE TABLE brands (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    website TEXT,
    CHECK (name <> ''),
    CHECK (website IS NULL OR website <> '')
);

CREATE TABLE skin_types (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    CHECK (name <> '')
);

CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price NUMERIC(10, 2) NOT NULL,
    category_id INT REFERENCES categories(id) ON DELETE SET NULL,
    brand_id INT REFERENCES brands(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CHECK (name <> ''),
    CHECK (price >= 0)
);

CREATE TABLE product_images (
    id SERIAL PRIMARY KEY,
    product_id INT REFERENCES products(id) ON DELETE CASCADE,
    image_url TEXT NOT NULL,
    alt_text TEXT,
    is_main BOOLEAN DEFAULT FALSE,
    CHECK (image_url <> '')
);

CREATE TABLE product_skin_types (
    product_id INT REFERENCES products(id) ON DELETE CASCADE,
    skin_type_id INT REFERENCES skin_types(id) ON DELETE CASCADE,
    PRIMARY KEY (product_id, skin_type_id)
);

CREATE INDEX idx_products_category_id ON products(category_id);
CREATE INDEX idx_products_brand_id ON products(brand_id);
CREATE INDEX idx_product_images_prod_id ON product_images(product_id);
CREATE UNIQUE INDEX idx_product_images_unique_main
    ON product_images(product_id)
    WHERE is_main;
CREATE INDEX idx_product_images_is_main ON product_images(is_main);
CREATE INDEX idx_product_skin_types_skin_type_id 
    ON product_skin_types(skin_type_id);
