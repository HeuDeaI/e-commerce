-- Begin transaction
BEGIN;

-- Insert categories (merged)
INSERT INTO categories (name, description) VALUES
  ('Очищение', 'Продукты для очищения кожи'),
  ('Тонизирование', 'Продукты для тонизирования кожи'),
  ('Кремы', 'Кремы для лица и тела'),
  ('Сыворотки', 'Сыворотки для ухода за кожей'),
  ('Маски', 'Маски для лица'),
  ('Солнцезащита', 'Солнцезащитные средства'),
  ('Антиоксиданты', 'Средства с антиоксидантами');

-- Insert brands (merged)
INSERT INTO brands (name, description, website) VALUES
  ('CERAVE', 'Косметика по уходу за кожей', 'https://www.cerave.com'),
  ('THE ORDINARY', 'Научная косметика', 'https://theordinary.com'),
  ('LA ROCHE-POSAY', 'Дерматологическая косметика', 'https://www.laroche-posay.com'),
  ('Bioderma', 'Медицинская косметика', 'https://www.bioderma.com'),
  ('Vichy', 'Уходовая косметика', 'https://www.vichy.com'),
  ('Eucerin', 'Средства для чувствительной кожи', 'https://www.eucerin.com');

-- Insert skin types (merged)
INSERT INTO skin_types (name, description) VALUES
  ('Нормальная', 'Нормальный тип кожи'),
  ('Сухая', 'Сухой тип кожи'),
  ('Жирная', 'Жирный тип кожи'),
  ('Комбинированная', 'Комбинированный тип кожи'),
  ('Чувствительная', 'Чувствительная кожа'),
  ('Стареющая', 'Кожа с признаками старения'),
  ('Проблемная', 'Кожа, склонная к акне'),
  ('Гиперчувствительная', 'Склонная к раздражению и аллергии'),
  ('Дерматитная', 'Кожа, склонная к дерматиту');

-- Insert products (merged)
INSERT INTO products (name, description, price, category_id, brand_id) VALUES
  ('Увлажняющий крем для лица', 'Крем для увлажнения лица', 42.35, 3, 1),
  ('Niacinamide 10% + Zinc 1% Serum', 'Сыворотка с ниацинамидом и цинком', 33.33, 2, 2),
  ('Витамин С сыворотка', 'Сыворотка с витамином С', 52.75, 2, 2),
  ('Антивозрастной крем', 'Крем против морщин', 58.95, 3, 3),
  ('Солнцезащитный SPF 50', 'Солнцезащитный крем', 31.50, 7, 4),
  ('Гиалуроновая кислота', 'Сыворотка с гиалуроновой кислотой', 45.75, 5, 2),
  ('Крем с ретинолом', 'Средство с ретинолом', 62.99, 6, 1);

-- Insert product images (merged)
INSERT INTO product_images (product_id, image_url, alt_text, is_main) VALUES
  (1, 'https://example.com/images/cream_main.jpg', 'Увлажняющий крем - главное изображение', TRUE),
  (1, 'https://example.com/images/cream_side.jpg', 'Увлажняющий крем - боковое изображение', FALSE),
  (2, 'https://example.com/images/serum_main.jpg', 'Сыворотка - главное изображение', TRUE),
  (4, 'https://example.com/images/antiage_main.jpg', 'Антивозрастной крем - главное изображение', TRUE),
  (5, 'https://example.com/images/sunscreen_main.jpg', 'Солнцезащитный крем - главное изображение', TRUE),
  (6, 'https://example.com/images/hyaluronic_main.jpg', 'Гиалуроновая сыворотка - главное изображение', TRUE),
  (7, 'https://example.com/images/retinol_main.jpg', 'Крем с ретинолом - главное изображение', TRUE);

-- Insert product skin types relationships (merged)
INSERT INTO product_skin_types (product_id, skin_type_id) VALUES
  (1, 1), (1, 2), (1, 5),
  (2, 4), (2, 3),
  (3, 1), (3, 3),
  (4, 6), (4, 7),
  (5, 8),
  (6, 4), (6, 5),
  (7, 1), (7, 2), (7, 6);

-- Commit transaction
COMMIT;
