package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/suprimkhatri77/turgorepo/backend/internal/models"
)

// TodoRepository handles todo persistence.
type TodoRepository struct {
	pool *pgxpool.Pool
}

// NewTodoRepository creates a new TodoRepository.
func NewTodoRepository(pool *pgxpool.Pool) *TodoRepository {
	return &TodoRepository{pool: pool}
}

// List returns todos with optional limit and offset.
func (r *TodoRepository) List(ctx context.Context, limit, offset int) ([]models.Todo, int, error) {
	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM todos`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	rows, err := r.pool.Query(ctx,
		`SELECT id, title, completed, created_at, updated_at FROM todos ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []models.Todo
	for rows.Next() {
		var t models.Todo
		if err := rows.Scan(&t.ID, &t.Title, &t.Completed, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, t)
	}
	return items, total, rows.Err()
}

// GetByID returns a todo by ID or pgx.ErrNoRows.
func (r *TodoRepository) GetByID(ctx context.Context, id uuid.UUID) (models.Todo, error) {
	var t models.Todo
	err := r.pool.QueryRow(ctx,
		`SELECT id, title, completed, created_at, updated_at FROM todos WHERE id = $1`,
		id).Scan(&t.ID, &t.Title, &t.Completed, &t.CreatedAt, &t.UpdatedAt)
	return t, err
}

// Create inserts a new todo and returns it.
func (r *TodoRepository) Create(ctx context.Context, input models.CreateTodoInput) (models.Todo, error) {
	id := uuid.New()
	var t models.Todo
	err := r.pool.QueryRow(ctx,
		`INSERT INTO todos (id, title, completed, created_at, updated_at)
		 VALUES ($1, $2, $3, NOW(), NOW())
		 RETURNING id, title, completed, created_at, updated_at`,
		id, input.Title, input.Completed).Scan(&t.ID, &t.Title, &t.Completed, &t.CreatedAt, &t.UpdatedAt)
	return t, err
}

// Update updates a todo by ID. Returns pgx.ErrNoRows if not found.
func (r *TodoRepository) Update(ctx context.Context, id uuid.UUID, input models.UpdateTodoInput) (models.Todo, error) {
	// Build dynamic update; for simplicity we do a read then write
	current, err := r.GetByID(ctx, id)
	if err != nil {
		return models.Todo{}, err
	}
	if input.Title != nil {
		current.Title = *input.Title
	}
	if input.Completed != nil {
		current.Completed = *input.Completed
	}
	err = r.pool.QueryRow(ctx,
		`UPDATE todos SET title = $2, completed = $3, updated_at = NOW()
		 WHERE id = $1
		 RETURNING id, title, completed, created_at, updated_at`,
		id, current.Title, current.Completed).Scan(&current.ID, &current.Title, &current.Completed, &current.CreatedAt, &current.UpdatedAt)
	return current, err
}

// Delete removes a todo by ID. Returns pgx.ErrNoRows if not found.
func (r *TodoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	cmd, err := r.pool.Exec(ctx, `DELETE FROM todos WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
