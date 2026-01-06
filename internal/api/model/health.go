package model

// HealthResponse represents the simplified health check response
// @name HealthResponse
type HealthResponse struct {
	Status string `json:"status" example:"ok"`
}
