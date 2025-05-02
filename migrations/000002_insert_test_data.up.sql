-- Begin transaction (optional, but recommended)
BEGIN;

-- Insert sample data for categories
INSERT INTO categories (name, description) VALUES
  ('Очищение', 'Продукты для очищения кожи'),
  ('Тонизирование', 'Продукты для тонизирования кожи'),
  ('Кремы', 'Кремы для лица и тела');

-- Insert sample data for brands
INSERT INTO brands (name, description, website) VALUES
  ('CERAVE', 'Косметика по уходу за кожей', 'https://www.cerave.com'),
  ('THE ORDINARY', 'Научная косметика', 'https://theordinary.com');

-- Insert sample data for skin types
INSERT INTO skin_types (name, description) VALUES
  ('Нормальная', 'Нормальный тип кожи'),
  ('Сухая', 'Сухой тип кожи'),
  ('Жирная', 'Жирный тип кожи'),
  ('Комбинированная', 'Комбинированный тип кожи'),
  ('Чувствительная', 'Чувствительная кожа');

-- Insert sample products.
-- Note: The integer values for category_id and brand_id correspond to the inserted rows above.
INSERT INTO products (name, description, price, category_id, brand_id) VALUES
  ('Увлажняющий крем для лица', 'Крем для увлажнения лица', 42.35, 3, 1),
  ('Niacinamide 10% + Zinc 1% Serum', 'Сыворотка с ниацинамидом и цинком', 33.33, 2, 2),
  ('Витамин С сыворотка', 'Сыворотка с витамином С', 52.75, 2, 2);

-- Insert sample product images
INSERT INTO product_images (product_id, image_url, alt_text, is_main) VALUES
  (1, 'https://example.com/images/cream_main.jpg', 'Увлажняющий крем - главное изображение', TRUE),
  (1, 'https://example.com/images/cream_side.jpg', 'Увлажняющий крем - боковое изображение', FALSE),
  (2, 'https://example.com/images/serum_main.jpg', 'Сыворотка - главное изображение', TRUE);

-- Insert sample many-to-many relationships for product_skin_types.
-- Mapping product 1 (Увлажняющий крем для лица) to skin types: Нормальная, Сухая, Чувствительная.
INSERT INTO product_skin_types (product_id, skin_type_id) VALUES
  (1, 1),
  (1, 2),
  (1, 5);
  
-- Mapping product 2 (Niacinamide 10% + Zinc 1% Serum) to skin types: Комбинированная, Жирная.
INSERT INTO product_skin_types (product_id, skin_type_id) VALUES
  (2, 4),
  (2, 3);
  
-- Mapping product 3 (Витамин С сыворотка) to skin types: Нормальная, Жирная.
INSERT INTO product_skin_types (product_id, skin_type_id) VALUES
  (3, 1),
  (3, 3);

-- Commit transaction to save changes
COMMIT;
