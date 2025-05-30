package domains

// Error represents an API error response
type Error struct {
	Message string `json:"error" example:"Error message"`
}
