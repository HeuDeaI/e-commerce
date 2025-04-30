package domains

import (
	"time"
)

type Brand struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Website     string    `json:"website"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Products    []Product `json:"products"`
}
