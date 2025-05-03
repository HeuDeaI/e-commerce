package domains

import (
	"time"
)

type Product struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Price       float64 `json:"price"`

	// Input/Write fields (for create/update operations)
	CategoryID  *int  `json:"category_id,omitempty"`
	BrandID     *int  `json:"brand_id,omitempty"`
	SkinTypeIDs []int `json:"skin_type_ids,omitempty"`

	// Output/Read fields (for response payloads)
	Category  *Category  `json:"category,omitempty"`
	Brand     *Brand     `json:"brand,omitempty"`
	SkinTypes []SkinType `json:"skin_types,omitempty"`

	// System fields
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

type ProductImage struct {
	ID        int    `json:"id"`
	ProductID int    `json:"product_id"`
	ImageURL  string `json:"image_url"`
	AltText   string `json:"alt_text,omitempty"`
	IsMain    bool   `json:"is_main"`
}
