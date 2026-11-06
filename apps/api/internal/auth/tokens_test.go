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

func testFamilyID() pgtype.UUID {
	id := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	return pgtype.UUID{Bytes: id, Valid: true}
}

func TestSignAndParseAccess(t *testing.T) {
	cfg := testConfig()
	user := testUser()

	token, err := SignAccess(cfg, user, uuid.New(), testFamilyID())
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
	if claims.SessionID != "22222222-2222-2222-2222-222222222222" {
		t.Fatalf("sid = %q", claims.SessionID)
	}
}

func TestParseAccessRejectsMissingSessionID(t *testing.T) {
	cfg := testConfig()
	claims := AccessClaims{
		UserID: "11111111-1111-1111-1111-111111111111",
		Role:   constants.RoleMember,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(cfg.JWTAccessSecret))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	if _, err := ParseAccess(signed, cfg.JWTAccessSecret); err == nil {
		t.Fatal("expected a token without sid to fail")
	}
}

func TestSignAccessSetsIssuedAt(t *testing.T) {
	cfg := testConfig()
	before := time.Now().Truncate(time.Second)

	token, err := SignAccess(cfg, testUser(), uuid.New(), testFamilyID())
	if err != nil {
		t.Fatalf("SignAccess: %v", err)
	}
	claims, err := ParseAccess(token, cfg.JWTAccessSecret)
	if err != nil {
		t.Fatalf("ParseAccess: %v", err)
	}

	if claims.IssuedAt == nil || claims.IssuedAt.Before(before) || claims.IssuedAt.After(time.Now()) {
		t.Fatalf("iat = %v", claims.IssuedAt)
	}
}

func TestIssuedAtOrBefore(t *testing.T) {
	cutoff := time.Unix(1_000, 500_000_000)

	tests := []struct {
		name     string
		issuedAt *jwt.NumericDate
		want     bool
	}{
		{"earlier", jwt.NewNumericDate(time.Unix(999, 0)), true},
		{"same second", jwt.NewNumericDate(time.Unix(1_000, 0)), true},
		{"later", jwt.NewNumericDate(time.Unix(1_001, 0)), false},
		{"no issue time", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := AccessClaims{RegisteredClaims: jwt.RegisteredClaims{IssuedAt: tt.issuedAt}}
			if got := claims.IssuedAtOrBefore(cutoff); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSignAndParseRefresh(t *testing.T) {
	cfg := testConfig()
	user := testUser()
	familyID := testFamilyID()

	token, err := SignRefresh(cfg, user, uuid.New(), familyID)
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
	if claims.FamilyID != "22222222-2222-2222-2222-222222222222" {
		t.Fatalf("family_id = %q, want 22222222-...", claims.FamilyID)
	}
}

func TestNewTokensIncludesFamilyID(t *testing.T) {
	cfg := testConfig()
	user := testUser()
	familyID := testFamilyID()

	tokens, err := NewTokens(cfg, user, familyID)
	if err != nil {
		t.Fatalf("NewTokens: %v", err)
	}

	claims, err := ParseRefresh(tokens.RefreshToken, cfg.JWTRefreshSecret)
	if err != nil {
		t.Fatalf("ParseRefresh: %v", err)
	}
	if claims.FamilyID != "22222222-2222-2222-2222-222222222222" {
		t.Fatalf("family_id = %q", claims.FamilyID)
	}
}

func TestNewAccessToken(t *testing.T) {
	cfg := testConfig()
	user := testUser()

	token, err := NewAccessToken(cfg, user, testFamilyID())
	if err != nil {
		t.Fatalf("NewAccessToken: %v", err)
	}

	claims, err := ParseAccess(token, cfg.JWTAccessSecret)
	if err != nil {
		t.Fatalf("ParseAccess: %v", err)
	}
	if claims.UserID != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("user_id = %q", claims.UserID)
	}
}

func TestParseRefreshRejectsMissingFamilyID(t *testing.T) {
	cfg := testConfig()
	// Manually sign a refresh token without FamilyID
	claims := RefreshClaims{
		UserID: "11111111-1111-1111-1111-111111111111",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(cfg.JWTRefreshSecret))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	if _, err := ParseRefresh(signed, cfg.JWTRefreshSecret); err == nil {
		t.Fatal("expected missing family_id to fail")
	}
}

func TestParseAccessRejectsWrongSecret(t *testing.T) {
	cfg := testConfig()
	token, err := SignAccess(cfg, testUser(), uuid.New(), testFamilyID())
	if err != nil {
		t.Fatalf("SignAccess: %v", err)
	}

	if _, err := ParseAccess(token, "other-secret"); err == nil {
		t.Fatal("expected wrong secret to fail")
	}
}

func TestParseAccessRejectsWrongAlg(t *testing.T) {
	claims := AccessClaims{
		UserID:    "11111111-1111-1111-1111-111111111111",
		SessionID: "22222222-2222-2222-2222-222222222222",
		Role:      constants.RoleMember,
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
	access, err := SignAccess(cfg, testUser(), uuid.New(), testFamilyID())
	if err != nil {
		t.Fatalf("SignAccess: %v", err)
	}

	if _, err := ParseRefresh(access, cfg.JWTRefreshSecret); err == nil {
		t.Fatal("expected access token to fail refresh parse")
	}
}

func TestHashTokenIsDeterministic(t *testing.T) {
	first := HashToken("refresh-token")
	second := HashToken("refresh-token")
	other := HashToken("other-token")

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
