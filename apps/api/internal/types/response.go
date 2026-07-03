package types

type AppError struct {
	Code    string `json:"code"`
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

type APIResponse struct {
	Success bool       `json:"success"`
	Message string     `json:"message,omitempty"`
	Code    string     `json:"code,omitempty"`
	Errors  []AppError `json:"errors,omitempty"`
	Data    any        `json:"data,omitempty"`
	Meta    any        `json:"meta,omitempty"`
}

type PaginationMeta struct {
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Offset     int `json:"offset"`
}
