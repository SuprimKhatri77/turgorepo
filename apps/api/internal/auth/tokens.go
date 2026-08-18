package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/suprimkhatri77/turgorepo/api/internal/config"
	db "github.com/suprimkhatri77/turgorepo/api/internal/database/generated"
	"github.com/suprimkhatri77/turgorepo/api/internal/utils"
)

type Tokens struct {
	AccessToken  string
	RefreshToken string
}

func HashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func NewTokens(cfg *config.Config, user db.User) (*Tokens, error) {
	jti := uuid.New()

	accessToken, err := SignAccess(cfg, user, jti)
	if err != nil {
		return nil, err
	}

	refreshToken, err := SignRefresh(cfg, user, jti)
	if err != nil {
		return nil, err
	}

	return &Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func SignAccess(cfg *config.Config, user db.User, jti uuid.UUID) (string, error) {
	userID, err := utils.UUIDString(user.ID)
	if err != nil {
		return "", err
	}

	claims := AccessClaims{
		UserID:   userID,
		Role:     user.Role,
		Email:    user.Email,
		Name:     user.Name,
		ImageURL: textValue(user.ImageUrl),
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.AccessTokenTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTAccessSecret))
}

func SignRefresh(cfg *config.Config, user db.User, jti uuid.UUID) (string, error) {
	userID, err := utils.UUIDString(user.ID)
	if err != nil {
		return "", err
	}

	claims := RefreshClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.RefreshTokenTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTRefreshSecret))
}

func ParseAccess(tokenString, secret string) (*AccessClaims, error) {
	claims := &AccessClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		hmacKeyFunc(secret),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		return nil, err
	}

	parsed, ok := token.Claims.(*AccessClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid access token claims")
	}
	if parsed.UserID == "" {
		return nil, fmt.Errorf("missing user_id")
	}
	return parsed, nil
}

func ParseRefresh(tokenString, secret string) (*RefreshClaims, error) {
	claims := &RefreshClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		hmacKeyFunc(secret),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		return nil, err
	}

	parsed, ok := token.Claims.(*RefreshClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid refresh token claims")
	}
	if parsed.UserID == "" {
		return nil, fmt.Errorf("missing user_id")
	}
	return parsed, nil
}

func hmacKeyFunc(secret string) jwt.Keyfunc {
	return func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	}
}

func textValue(value pgtype.Text) string {
	if !value.Valid {
		return ""
	}
	return value.String
}
