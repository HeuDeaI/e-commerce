package domains

import (
	"time"
)

type Product struct {
	ID          int            `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Price       float64        `json:"price"`
	CategoryID  *int           `json:"category_id"`
	Category    *Category      `json:"category"`
	BrandID     *int           `json:"brand_id"`
	Brand       *Brand         `json:"brand"`
	SkinTypes   []SkinType     `json:"skin_type"`
	Images      []ProductImage `json:"images"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type ProductImage struct {
	ID        int    `json:"id"`
	ProductID int    `json:"product_id"`
	ImageURL  string `json:"image_url"`
	AltText   string `json:"alt_text"`
	IsMain    bool   `json:"is_main"`
}
