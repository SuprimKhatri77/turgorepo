package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/suprimkhatri77/turgorepo/api/internal/config"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	db "github.com/suprimkhatri77/turgorepo/api/internal/database/generated"
	"github.com/suprimkhatri77/turgorepo/api/internal/response"
	"github.com/suprimkhatri77/turgorepo/api/internal/rlog"
	"github.com/suprimkhatri77/turgorepo/api/internal/session"
	"github.com/suprimkhatri77/turgorepo/api/internal/utils"
)

// Refresh exchanges the refresh token cookie for new tokens. The presented token
// is one of three things:
//   - the session's current token: rotate it and issue a new pair
//   - the previous token, rotated moments ago: a client that sent two refreshes
//     at once, so issue an access token only
//   - anything older: a reused, possibly stolen token, so end the session
func (h *Handler) Refresh(c *gin.Context) {
	ctx := c.Request.Context()

	refreshTokenString, err := c.Cookie(session.RefreshTokenCookie)
	if err != nil {
		rlog.Warn(c, "missing refresh token cookie")
		rejectSession(c, h.Config, constants.TokenNotProvided, "Missing refresh token")
		return
	}

	claims, err := session.ParseRefresh(refreshTokenString, h.Config.JWTRefreshSecret)
	if err != nil {
		rlog.Warn(c, "invalid refresh token", "error", err)
		rejectSession(c, h.Config, constants.InvalidToken, "Invalid refresh token")
		return
	}

	familyID, err := utils.ConvertToUUID(claims.FamilyID)
	if err != nil {
		rlog.Warn(c, "invalid family_id in refresh token", "error", err)
		rejectSession(c, h.Config, constants.InvalidToken, "Invalid token claims")
		return
	}

	sess, err := h.Store.GetActiveSessionByID(ctx, familyID)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			respondSessionLookupFailed(c, err)
			return
		}
		rlog.Warn(c, "session not found or expired", "family_id", claims.FamilyID)
		rejectSession(c, h.Config, constants.InvalidToken, "Session expired")
		return
	}

	user, err := h.Store.GetUserByID(ctx, sess.UserID)
	if err != nil {
		rlog.Error(c, "failed to fetch user", err, "user_id", sess.UserID)
		response.Internal(c)
		return
	}

	presentedHash := session.HashToken(refreshTokenString)

	switch {
	case presentedHash == sess.CurrentTokenHash:
		h.rotate(c, user, sess, presentedHash)
	case withinGrace(sess, presentedHash, h.Config.RefreshReuseWindow, time.Now()):
		h.issueGraceAccess(c, user, familyID)
	default:
		h.endReusedSession(c, user, sess, claims.FamilyID)
	}
}

// rotate issues a new token pair and swaps the session's current token for it.
// The swap only succeeds if the session still holds presentedHash, so two
// concurrent refreshes can't both rotate.
func (h *Handler) rotate(c *gin.Context, user db.User, sess db.Session, presentedHash string) {
	tokens, err := session.NewTokens(h.Config, user, sess.ID)
	if err != nil {
		rlog.Error(c, "failed to sign tokens", err, "user_id", user.ID)
		response.Internal(c)
		return
	}

	_, err = h.Store.RotateSessionToken(c.Request.Context(), db.RotateSessionTokenParams{
		ID:                sess.ID,
		NewTokenHash:      session.HashToken(tokens.RefreshToken),
		ExpectedTokenHash: presentedHash,
		ExpiresAt:         pgtype.Timestamptz{Time: time.Now().Add(h.Config.RefreshTokenTTL), Valid: true},
	})
	if errors.Is(err, pgx.ErrNoRows) {
		h.handleLostRace(c, user, sess.ID, presentedHash)
		return
	}
	if err != nil {
		rlog.Error(c, "failed to rotate session token", err, "user_id", user.ID)
		response.Internal(c)
		return
	}

	session.SetCookies(c, h.Config, tokens.AccessToken, tokens.RefreshToken)
	rlog.Info(c, "tokens rotated successfully", "user_id", user.ID)
	response.OK(c, http.StatusOK, "Tokens refreshed", nil)
}

// handleLostRace runs when another request rotated the token between this
// request's read and its swap. That request now holds the new pair, so this one
// is treated as presenting the previous token: reload the session and serve the
// grace path instead of failing the client.
func (h *Handler) handleLostRace(c *gin.Context, user db.User, familyID pgtype.UUID, presentedHash string) {
	reloaded, err := h.Store.GetActiveSessionByID(c.Request.Context(), familyID)
	if errors.Is(err, pgx.ErrNoRows) {
		rlog.Warn(c, "session gone after lost rotation race")
		rejectSession(c, h.Config, constants.InvalidToken, "Session expired")
		return
	}
	if err != nil {
		respondSessionLookupFailed(c, err)
		return
	}

	if !withinGrace(reloaded, presentedHash, h.Config.RefreshReuseWindow, time.Now()) {
		rlog.Warn(c, "lost rotation race but grace window missed", "user_id", user.ID)
		rejectSession(c, h.Config, constants.InvalidToken, "Session invalid, please log in again")
		return
	}
	h.issueGraceAccess(c, user, familyID)
}

// issueGraceAccess answers a refresh with the previous token: a new access token,
// but no new refresh token, since the client already received one.
func (h *Handler) issueGraceAccess(c *gin.Context, user db.User, familyID pgtype.UUID) {
	accessToken, err := session.NewAccessToken(h.Config, user, familyID)
	if err != nil {
		rlog.Error(c, "failed to sign access token", err, "user_id", user.ID)
		response.Internal(c)
		return
	}

	session.SetAccessCookie(c, h.Config, accessToken)
	rlog.Info(c, "grace-window refresh (previous token reuse)", "user_id", user.ID)
	response.OK(c, http.StatusOK, "Tokens refreshed", nil)
}

// endReusedSession handles a refresh token that is neither current nor recently
// rotated. It may have been stolen, so end the session, including access tokens
// already issued for it, whoever holds them.
func (h *Handler) endReusedSession(c *gin.Context, user db.User, sess db.Session, sessionID string) {
	rlog.Warn(c, "stale or reused token outside grace window", "user_id", user.ID)
	if err := h.Store.RevokeSession(c.Request.Context(), sess.ID); err != nil {
		rlog.Error(c, "failed to revoke session after stale token", err)
	}
	h.revokeSessionTokens(c, sessionID)
	rejectSession(c, h.Config, constants.InvalidToken, "Session invalid, please log in again")
}

// withinGrace reports whether presentedHash is the session's previous token and
// it was rotated less than window before now.
func withinGrace(s db.Session, presentedHash string, window time.Duration, now time.Time) bool {
	return s.PreviousTokenHash.Valid &&
		presentedHash == s.PreviousTokenHash.String &&
		s.PreviousRotatedAt.Valid &&
		now.Sub(s.PreviousRotatedAt.Time) < window
}

// rejectSession clears the session cookies and answers 401, so the client
// treats the user as logged out.
func rejectSession(c *gin.Context, cfg *config.Config, code, message string) {
	session.ClearCookies(c, cfg)
	response.Error(c, http.StatusUnauthorized, code, message)
}

// respondSessionLookupFailed reports a database failure without clearing cookies:
// the session may still be valid, so an outage must not log the user out.
func respondSessionLookupFailed(c *gin.Context, err error) {
	rlog.Error(c, "failed to look up session", err)
	response.Error(c, http.StatusServiceUnavailable, constants.ServiceUnavailable, "Service temporarily unavailable")
}
