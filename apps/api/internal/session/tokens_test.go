package session

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)

	claims, err := ParseAccess(token, cfg.JWTAccessSecret)
	require.NoError(t, err)

	require.Equal(t, "11111111-1111-1111-1111-111111111111", claims.UserID)
	require.Equal(t, constants.RoleMember, claims.Role)
	require.Equal(t, user.Email, claims.Email)
	require.Equal(t, "22222222-2222-2222-2222-222222222222", claims.SessionID)
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
	require.NoError(t, err)

	_, err = ParseAccess(signed, cfg.JWTAccessSecret)
	require.Error(t, err, "expected a token without sid to fail")
}

func TestSignAccessSetsIssuedAt(t *testing.T) {
	cfg := testConfig()
	before := time.Now().Truncate(time.Second)

	token, err := SignAccess(cfg, testUser(), uuid.New(), testFamilyID())
	require.NoError(t, err)
	claims, err := ParseAccess(token, cfg.JWTAccessSecret)
	require.NoError(t, err)

	require.NotNil(t, claims.IssuedAt)
	require.WithinRange(t, claims.IssuedAt.Time, before, time.Now())
}

func TestSignAndParseRefresh(t *testing.T) {
	cfg := testConfig()
	user := testUser()
	familyID := testFamilyID()

	token, err := SignRefresh(cfg, user, uuid.New(), familyID)
	require.NoError(t, err)

	claims, err := ParseRefresh(token, cfg.JWTRefreshSecret)
	require.NoError(t, err)

	require.Equal(t, "11111111-1111-1111-1111-111111111111", claims.UserID)
	require.Equal(t, "22222222-2222-2222-2222-222222222222", claims.FamilyID)
}

func TestNewTokensIncludesFamilyID(t *testing.T) {
	cfg := testConfig()
	user := testUser()
	familyID := testFamilyID()

	tokens, err := NewTokens(cfg, user, familyID)
	require.NoError(t, err)

	claims, err := ParseRefresh(tokens.RefreshToken, cfg.JWTRefreshSecret)
	require.NoError(t, err)
	require.Equal(t, "22222222-2222-2222-2222-222222222222", claims.FamilyID)
}

func TestNewAccessToken(t *testing.T) {
	cfg := testConfig()
	user := testUser()

	token, err := NewAccessToken(cfg, user, testFamilyID())
	require.NoError(t, err)

	claims, err := ParseAccess(token, cfg.JWTAccessSecret)
	require.NoError(t, err)
	require.Equal(t, "11111111-1111-1111-1111-111111111111", claims.UserID)
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
	require.NoError(t, err)

	_, err = ParseRefresh(signed, cfg.JWTRefreshSecret)
	require.Error(t, err, "expected missing family_id to fail")
}

func TestParseAccessRejectsWrongSecret(t *testing.T) {
	cfg := testConfig()
	token, err := SignAccess(cfg, testUser(), uuid.New(), testFamilyID())
	require.NoError(t, err)

	_, err = ParseAccess(token, "other-secret")
	require.Error(t, err, "expected wrong secret to fail")
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
	require.NoError(t, err)

	_, err = ParseAccess(signed, "access-secret")
	require.Error(t, err, "expected HS384 token to fail")
}

func TestParseRefreshRejectsAccessToken(t *testing.T) {
	cfg := testConfig()
	access, err := SignAccess(cfg, testUser(), uuid.New(), testFamilyID())
	require.NoError(t, err)

	_, err = ParseRefresh(access, cfg.JWTRefreshSecret)
	require.Error(t, err, "expected access token to fail refresh parse")
}

func TestHashTokenIsDeterministic(t *testing.T) {
	first := HashToken("refresh-token")
	second := HashToken("refresh-token")
	other := HashToken("other-token")

	require.Equal(t, first, second, "hash should be deterministic")
	require.NotEqual(t, first, other, "different tokens should hash differently")
	require.Len(t, first, 64, "sha256 hex length")
}

func TestEmailTokensAreRandomAndURLSafe(t *testing.T) {
	first, firstHash, err := NewEmailToken()
	require.NoError(t, err)
	second, _, err := NewEmailToken()
	require.NoError(t, err)

	assert.NotEqual(t, first, second, "tokens repeat")
	assert.Len(t, first, 43, "32 bytes")
	assert.False(t, strings.ContainsAny(first, "+/="), "token isn't URL safe: %s", first)
	assert.Equal(t, HashToken(first), firstHash, "hash doesn't match")
}
