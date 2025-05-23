package domains

import (
	"time"
)

type ProductRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Price       float64 `json:"price"`
	CategoryID  *int    `json:"category_id,omitempty"`
	BrandID     *int    `json:"brand_id,omitempty"`
	SkinTypeIDs []int   `json:"skin_type_ids,omitempty"`
}

type ProductResponse struct {
	ID          int        `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Price       float64    `json:"price"`
	Category    *Category  `json:"category,omitempty"`
	Brand       *Brand     `json:"brand,omitempty"`
	SkinTypes   []SkinType `json:"skin_types,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

type ProductImage struct {
	ID        int    `json:"id"`
	ProductID int    `json:"product_id"`
	ImageURL  string `json:"image_url"`
	AltText   string `json:"alt_text,omitempty"`
	IsMain    bool   `json:"is_main"`
	ImageData string `json:"image_data,omitempty"`
}

type PriceRange struct {
	MinPrice *float64 `json:"min_price,omitempty"`
	MaxPrice *float64 `json:"max_price,omitempty"`
}
