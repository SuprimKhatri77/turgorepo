// Package ratelimit limits how often a client can call sensitive endpoints.
package ratelimit

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/suprimkhatri77/turgorepo/api/internal/constants"
	"github.com/suprimkhatri77/turgorepo/api/internal/packages/rlog"
	"github.com/suprimkhatri77/turgorepo/api/internal/types"
)

// checkTimeout keeps a slow Redis from delaying auth requests.
const checkTimeout = 500 * time.Millisecond

// Limiter counts hits against a key within a fixed window.
type Limiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (allowed bool, retryAfter time.Duration, err error)
}

// Rule is a named request budget.
type Rule struct {
	Name   string
	Limit  int
	Window time.Duration
}

var (
	// LoginPerIP slows guessing across many accounts from one address.
	LoginPerIP = Rule{Name: "login:ip", Limit: 20, Window: time.Minute}
	// LoginPerAccount slows guessing one account's password.
	LoginPerAccount = Rule{Name: "login:account", Limit: 10, Window: 15 * time.Minute}
	// RegisterPerIP slows bulk account creation.
	RegisterPerIP = Rule{Name: "register:ip", Limit: 10, Window: time.Hour}
)

// ByIP returns middleware that applies rule to each client IP.
func ByIP(limiter Limiter, rule Rule) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !Check(c, limiter, rule, c.ClientIP()) {
			c.Abort()
			return
		}
		c.Next()
	}
}

// AccountKey identifies one client guessing at one account. The email is hashed
// to keep addresses out of Redis.
func AccountKey(c *gin.Context, email string) string {
	sum := sha256.Sum256([]byte(email))
	return c.ClientIP() + ":" + hex.EncodeToString(sum[:])
}

// Check applies rule to key. If the budget is spent it writes a 429 and returns
// false. If the limiter fails it logs the error and allows the request, so a
// Redis outage doesn't lock everyone out of logging in.
func Check(c *gin.Context, limiter Limiter, rule Rule, key string) bool {
	ctx, cancel := context.WithTimeout(c.Request.Context(), checkTimeout)
	defer cancel()

	allowed, retryAfter, err := limiter.Allow(ctx, "ratelimit:"+rule.Name+":"+key, rule.Limit, rule.Window)
	if err != nil {
		rlog.Error(c, "rate limit check failed, allowing request", err, "rule", rule.Name)
		return true
	}
	if allowed {
		return true
	}

	rlog.Warn(c, "rate limit exceeded", "rule", rule.Name)
	c.Header("Retry-After", strconv.Itoa(int(math.Ceil(retryAfter.Seconds()))))
	c.JSON(http.StatusTooManyRequests, types.APIResponse{
		Success: false,
		Message: "Too many attempts, please try again later",
		Code:    constants.RateLimited,
	})
	return false
}
