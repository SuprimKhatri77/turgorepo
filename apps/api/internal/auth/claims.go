package auth

import "github.com/golang-jwt/jwt/v5"

type AccessClaims struct {
	UserID   string `json:"user_id"`
	Role     string `json:"role"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	ImageURL string `json:"image_url,omitempty"`
	jwt.RegisteredClaims
}

type RefreshClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}
