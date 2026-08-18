package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/suprimkhatri77/turgorepo/api/internal/config"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	db "github.com/suprimkhatri77/turgorepo/api/internal/database/generated"
)

func testConfig() *config.Config {
	return &config.Config{
		JWTAccessSecret:    "access-secret",
		JWTRefreshSecret:   "refresh-secret",
		AccessTokenTTL:     15 * time.Minute,
		RefreshTokenTTL:    30 * 24 * time.Hour,
		RefreshReuseWindow: 5 * time.Minute,
	}
}

func testUser() db.User {
	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	return db.User{
		ID:    pgtype.UUID{Bytes: id, Valid: true},
		Name:  "Ada Lovelace",
		Email: "ada@example.com",
		Role:  constants.RoleMember,
	}
}

func TestSignAndParseAccess(t *testing.T) {
	cfg := testConfig()
	user := testUser()

	token, err := SignAccess(cfg, user, uuid.New())
	if err != nil {
		t.Fatalf("SignAccess: %v", err)
	}

	claims, err := ParseAccess(token, cfg.JWTAccessSecret)
	if err != nil {
		t.Fatalf("ParseAccess: %v", err)
	}

	if claims.UserID != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("user_id = %q", claims.UserID)
	}
	if claims.Role != constants.RoleMember {
		t.Fatalf("role = %q", claims.Role)
	}
	if claims.Email != user.Email {
		t.Fatalf("email = %q", claims.Email)
	}
}

func TestSignAndParseRefresh(t *testing.T) {
	cfg := testConfig()
	user := testUser()

	token, err := SignRefresh(cfg, user, uuid.New())
	if err != nil {
		t.Fatalf("SignRefresh: %v", err)
	}

	claims, err := ParseRefresh(token, cfg.JWTRefreshSecret)
	if err != nil {
		t.Fatalf("ParseRefresh: %v", err)
	}

	if claims.UserID != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("user_id = %q", claims.UserID)
	}
}

func TestParseAccessRejectsWrongSecret(t *testing.T) {
	cfg := testConfig()
	token, err := SignAccess(cfg, testUser(), uuid.New())
	if err != nil {
		t.Fatalf("SignAccess: %v", err)
	}

	if _, err := ParseAccess(token, "other-secret"); err == nil {
		t.Fatal("expected wrong secret to fail")
	}
}

func TestParseAccessRejectsWrongAlg(t *testing.T) {
	claims := AccessClaims{
		UserID: "11111111-1111-1111-1111-111111111111",
		Role:   constants.RoleMember,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)
	signed, err := token.SignedString([]byte("access-secret"))
	if err != nil {
		t.Fatalf("sign HS384: %v", err)
	}

	if _, err := ParseAccess(signed, "access-secret"); err == nil {
		t.Fatal("expected HS384 token to fail")
	}
}

func TestParseRefreshRejectsAccessToken(t *testing.T) {
	cfg := testConfig()
	access, err := SignAccess(cfg, testUser(), uuid.New())
	if err != nil {
		t.Fatalf("SignAccess: %v", err)
	}

	if _, err := ParseRefresh(access, cfg.JWTRefreshSecret); err == nil {
		t.Fatal("expected access token to fail refresh parse")
	}
}

func TestHashRefreshTokenIsDeterministic(t *testing.T) {
	first := HashRefreshToken("refresh-token")
	second := HashRefreshToken("refresh-token")
	other := HashRefreshToken("other-token")

	if first != second {
		t.Fatal("hash should be deterministic")
	}
	if first == other {
		t.Fatal("different tokens should hash differently")
	}
	if len(first) != 64 {
		t.Fatalf("sha256 hex length = %d", len(first))
	}
}
