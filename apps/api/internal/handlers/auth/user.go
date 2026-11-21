package auth

import (
	"github.com/google/uuid"
	db "github.com/suprimkhatri77/turgorepo/api/internal/database/generated"
)

// userResponse is the user shape returned by the API. It matches UserSchema in
// @repo/types.
type userResponse struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Email    string  `json:"email"`
	Role     string  `json:"role"`
	ImageURL *string `json:"image_url,omitempty"`
}

// newUserResponse keeps internal fields like the password hash out of responses.
func newUserResponse(user db.User) userResponse {
	public := userResponse{
		ID:    uuid.UUID(user.ID.Bytes).String(),
		Name:  user.Name,
		Email: user.Email,
		Role:  user.Role,
	}
	if user.ImageUrl.Valid {
		public.ImageURL = &user.ImageUrl.String
	}
	return public
}
