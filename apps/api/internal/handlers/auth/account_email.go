package auth

import (
	"context"

	session "github.com/suprimkhatri77/turgorepo/api/internal/auth"
	"github.com/suprimkhatri77/turgorepo/api/internal/config"
	"github.com/suprimkhatri77/turgorepo/api/internal/packages/email"
	"github.com/suprimkhatri77/turgorepo/api/internal/ratelimit"
	"github.com/suprimkhatri77/turgorepo/api/internal/repository"
)

// JobRunner runs work after the response has been sent.
type JobRunner interface {
	Go(name string, job func(ctx context.Context))
}

// AccountEmailDeps are the dependencies of the handlers that send account email:
// registration, email verification and password reset.
type AccountEmailDeps struct {
	Queries repository.AuthRepository
	Config  *config.Config
	Limiter ratelimit.Limiter
	Revoker session.TokenRevoker
	Mailer  email.Mailer
	Jobs    JobRunner
}
