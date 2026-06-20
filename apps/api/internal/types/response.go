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
	Data    any        `json:"data"`
	Meta    any        `json:"meta,omitempty"`
}

type PaginationMeta struct {
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
	Page       int `json:"page,omitempty"`
	Limit      int `json:"limit"`
}
