package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/suprimkhatri77/turgorepo/backend/internal/errors"
	"github.com/suprimkhatri77/turgorepo/backend/internal/models"
	"github.com/suprimkhatri77/turgorepo/backend/internal/repository"
)

// TodoHandler handles todo HTTP requests.
type TodoHandler struct {
	repo *repository.TodoRepository
}

// NewTodoHandler creates a new TodoHandler.
func NewTodoHandler(repo *repository.TodoRepository) *TodoHandler {
	return &TodoHandler{repo: repo}
}

// List returns paginated todos.
func (h *TodoHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	items, total, err := h.repo.List(c.Request.Context(), limit, offset)
	if err != nil {
		JSONError(c, errors.Wrap(errors.ErrInternal, err))
		return
	}
	JSON(c, http.StatusOK, gin.H{"items": items, "total": total})
}

// GetByID returns a single todo by ID.
func (h *TodoHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		JSONError(c, errors.WithMessage(errors.ErrBadRequest, "invalid id"))
		return
	}
	todo, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == pgx.ErrNoRows {
			JSONError(c, errors.ErrNotFound)
			return
		}
		JSONError(c, errors.Wrap(errors.ErrInternal, err))
		return
	}
	JSON(c, http.StatusOK, todo)
}

// Create creates a new todo.
func (h *TodoHandler) Create(c *gin.Context) {
	var input models.CreateTodoInput
	if !BindJSON(c, &input) {
		return
	}
	todo, err := h.repo.Create(c.Request.Context(), input)
	if err != nil {
		JSONError(c, errors.Wrap(errors.ErrInternal, err))
		return
	}
	JSON(c, http.StatusCreated, todo)
}

// Update updates an existing todo.
func (h *TodoHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		JSONError(c, errors.WithMessage(errors.ErrBadRequest, "invalid id"))
		return
	}
	var input models.UpdateTodoInput
	if !BindJSON(c, &input) {
		return
	}
	todo, err := h.repo.Update(c.Request.Context(), id, input)
	if err != nil {
		if err == pgx.ErrNoRows {
			JSONError(c, errors.ErrNotFound)
			return
		}
		JSONError(c, errors.Wrap(errors.ErrInternal, err))
		return
	}
	JSON(c, http.StatusOK, todo)
}

// Delete deletes a todo.
func (h *TodoHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		JSONError(c, errors.WithMessage(errors.ErrBadRequest, "invalid id"))
		return
	}
	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		if err == pgx.ErrNoRows {
			JSONError(c, errors.ErrNotFound)
			return
		}
		JSONError(c, errors.Wrap(errors.ErrInternal, err))
		return
	}
	c.Status(http.StatusNoContent)
}
