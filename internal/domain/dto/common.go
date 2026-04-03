package dto

type SuccessResponse[T any] struct {
	Data    T      `json:"data"`
	Message string `json:"message,omitempty"`
}

type PaginatedResponse[T any] struct {
	Data       []T   `json:"data"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"total_pages"`
}
