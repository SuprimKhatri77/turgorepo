package models

import (
	"time"

	"github.com/google/uuid"
)

// Todo represents a todo item in the database and API.
type Todo struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateTodoInput is the input for creating a todo.
type CreateTodoInput struct {
	Title     string `json:"title" binding:"required,min=1"`
	Completed bool   `json:"completed"`
}

// UpdateTodoInput is the input for updating a todo (all fields optional).
type UpdateTodoInput struct {
	Title     *string `json:"title"`
	Completed *bool   `json:"completed"`
}
