package domains

import (
	"time"
)

type Product struct {
	ID          int            `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Price       float64        `json:"price"`
	CategoryID  *int           `json:"category_id,omitempty"`
	Category    *Category      `json:"category,omitempty"`
	BrandID     *int           `json:"brand_id,omitempty"`
	Brand       *Brand         `json:"brand,omitempty"`
	SkinTypes   []SkinType     `json:"skin_type,omitempty"`
	Images      []ProductImage `json:"images,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"-"`
}

type ProductImage struct {
	ID        int    `json:"id"`
	ProductID int    `json:"product_id"`
	ImageURL  string `json:"image_url"`
	AltText   string `json:"alt_text,omitempty"`
	IsMain    bool   `json:"is_main"`
}
