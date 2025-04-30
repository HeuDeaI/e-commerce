package domains

import (
	"time"
)

type Brand struct {
	ID          int       `json:"id,omitempty"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Website     string    `json:"website,omitempty"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`
	Products    []Product `json:"products,omitempty"`
}
