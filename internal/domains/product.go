package domains

import (
	"time"
)

type Product struct {
	ID          uint           `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Price       float64        `json:"price"`
	CategoryID  *uint          `json:"category_id"`
	Category    *Category      `json:"category"`
	BrandID     *uint          `json:"brand_id"`
	Brand       *Brand         `json:"brand"`
	Images      []ProductImage `json:"images"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type ProductImage struct {
	ID        uint   `json:"id"`
	ProductID uint   `json:"product_id"`
	ImageURL  string `json:"image_url"`
	AltText   string `json:"alt_text"`
	IsMain    bool   `json:"is_main"`
}
