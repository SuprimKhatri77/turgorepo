package auth

import (
	"context"
	"log/slog"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/suprimkhatri77/turgorepo/api/internal/config"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	"github.com/suprimkhatri77/turgorepo/api/internal/email"
	"github.com/suprimkhatri77/turgorepo/api/internal/ratelimit"
	"github.com/suprimkhatri77/turgorepo/api/internal/response"
	"github.com/suprimkhatri77/turgorepo/api/internal/rlog"
	"github.com/suprimkhatri77/turgorepo/api/internal/session"
	"github.com/suprimkhatri77/turgorepo/api/internal/utils"
)

// Handler serves the /auth routes. Build one in routes and register its
// methods, e.g. r.POST("/login", h.Login).
type Handler struct {
	Store   Store
	Config  *config.Config
	Limiter ratelimit.Limiter
	Revoker session.TokenRevoker
	Mailer  email.Mailer
	Jobs    JobRunner
}

// JobRunner runs work after the response has been sent.
type JobRunner interface {
	Go(name string, job func(ctx context.Context))
}

// sendInBackground sends an email after the response has gone out. A failure is
// only logged, because the request itself has already succeeded.
func (h *Handler) sendInBackground(c *gin.Context, name string, send func(ctx context.Context) error) {
	requestID := c.GetString(constants.RequestIDKey)
	h.Jobs.Go(name, func(ctx context.Context) {
		if err := send(ctx); err != nil {
			slog.Error("failed to send "+name, "request_id", requestID, "error", err)
		}
	})
}

// frontendLink builds a link to a web app page that carries token. It uses
// configuration rather than request headers, so a forged Host can't send the
// token to someone else's site. The token is in the fragment because browsers
// don't send fragments to servers, which keeps it out of access logs.
func frontendLink(cfg *config.Config, path, token string) string {
	return strings.TrimRight(cfg.FrontendURL, "/") + path + "#token=" + token
}

// currentUserID returns the user ID that RequireAuth stored for this request, as
// text and as a UUID. A missing or malformed ID is a server bug rather than a bad
// request, so it writes a 500 and returns false.
func currentUserID(c *gin.Context) (string, pgtype.UUID, bool) {
	text := c.GetString(constants.UserIDKey)
	id, err := utils.ConvertToUUID(text)
	if err != nil {
		rlog.Error(c, "invalid user_id in context", err, "user_id", text)
		response.Internal(c)
		return "", pgtype.UUID{}, false
	}
	return text, id, true
}
