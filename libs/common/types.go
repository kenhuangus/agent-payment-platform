package common

import (
	"time"
)

// BaseEntity represents common fields for all entities
type BaseEntity struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// APIResponse represents a standard API response structure
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// APIError represents an API error response
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Meta represents pagination and metadata for list responses
type Meta struct {
	Page       int `json:"page,omitempty"`
	Limit      int `json:"limit,omitempty"`
	Total      int `json:"total,omitempty"`
	TotalPages int `json:"totalPages,omitempty"`
}

// ListResponse represents a paginated list response
type ListResponse struct {
	Items []interface{} `json:"items"`
	Meta  Meta          `json:"meta"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

// NewSuccessResponse creates a successful API response
func NewSuccessResponse(data interface{}) *APIResponse {
	return &APIResponse{
		Success: true,
		Data:    data,
	}
}

// NewErrorResponse creates an error API response
func NewErrorResponse(code, message string) *APIResponse {
	return &APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
		},
	}
}

// NewErrorResponseWithDetails creates an error API response with details
func NewErrorResponseWithDetails(code, message, details string) *APIResponse {
	return &APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}

// NewValidationErrorResponse creates a validation error response
func NewValidationErrorResponse(errors []ValidationError) *APIResponse {
	return &APIResponse{
		Success: false,
		Error: &APIError{
			Code:    "VALIDATION_ERROR",
			Message: "Validation failed",
		},
		Data: ValidationErrors{Errors: errors},
	}
}

// NewListResponse creates a paginated list response
func NewListResponse(items []interface{}, page, limit, total int) *ListResponse {
	totalPages := (total + limit - 1) / limit // Ceiling division
	return &ListResponse{
		Items: items,
		Meta: Meta{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}
}
