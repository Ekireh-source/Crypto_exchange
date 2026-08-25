package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"

	"exchange/internal/models"
	"exchange/internal/services"
)

// RequireAPIKey validates the API key from the Authorization header or X-API-Key header.
// On success, it injects the APIKeyContext and the owner's userID into echo.Context,
// so existing handlers that call middleware.GetUserID(c) work unchanged.
func RequireAPIKey(apiKeySvc *services.APIKeyService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			rawKey := extractAPIKey(c)
			if rawKey == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "API key required. Pass via Authorization: Bearer sk_live_xxx or X-API-Key header.")
			}

			keyCtx, err := apiKeySvc.ValidateKey(c.Request().Context(), rawKey)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or revoked API key")
			}

			// Enforce: publishable keys (pk_) can only call GET endpoints
			if keyCtx.KeyType == models.APIKeyTypePublishable && c.Request().Method != http.MethodGet {
				return echo.NewHTTPError(http.StatusForbidden,
					"This endpoint requires a secret key (sk_). Publishable keys (pk_) are read-only.")
			}

			// Inject into context — same key as JWT middleware uses
			c.Set("apiKeyCtx", keyCtx)
			c.Set("userID", keyCtx.OwnerUserID)

			return next(c)
		}
	}
}

// extractAPIKey reads the raw key from either Authorization or X-API-Key header.
func extractAPIKey(c echo.Context) string {
	// Try Authorization: Bearer sk_live_xxx
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return strings.TrimSpace(parts[1])
		}
	}

	// Try X-API-Key: sk_live_xxx
	return strings.TrimSpace(c.Request().Header.Get("X-API-Key"))
}

// ── Rate Limiter ──────────────────────────────────────────────────────────────

var (
	limiters   = sync.Map{} // map[uuid.UUID]*rate.Limiter
	defaultRPM = 60         // requests per minute
)

// RateLimitByAPIKey limits requests per API key using a token bucket.
func RateLimitByAPIKey() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			keyCtx, ok := c.Get("apiKeyCtx").(*models.APIKeyContext)
			if !ok || keyCtx == nil {
				return next(c)
			}

			limiter := getOrCreateLimiter(keyCtx.KeyID)
			if !limiter.Allow() {
				return echo.NewHTTPError(http.StatusTooManyRequests,
					"Rate limit exceeded. Maximum 60 requests per minute.")
			}

			return next(c)
		}
	}
}

func getOrCreateLimiter(keyID uuid.UUID) *rate.Limiter {
	if v, ok := limiters.Load(keyID); ok {
		return v.(*rate.Limiter)
	}

	// rate.NewLimiter(r, b): r = tokens per second, b = burst size
	// 60 req/min = 1 req/sec, with burst of 10
	l := rate.NewLimiter(rate.Every(time.Minute/time.Duration(defaultRPM)), 10)
	limiters.Store(keyID, l)
	return l
}

// ── Request Logger ────────────────────────────────────────────────────────────

// APIRequestLogger logs every API key request asynchronously.
func APIRequestLogger(apiKeySvc *services.APIKeyService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			// Execute the handler
			err := next(c)

			// Log after response is sent
			keyCtx, ok := c.Get("apiKeyCtx").(*models.APIKeyContext)
			if ok && keyCtx != nil {
				latency := int(time.Since(start).Milliseconds())
				apiKeySvc.LogRequest(c.Request().Context(), &models.APIRequestLog{
					KeyID:      keyCtx.KeyID,
					Method:     c.Request().Method,
					Path:       c.Request().URL.Path,
					StatusCode: c.Response().Status,
					LatencyMs:  latency,
				})
				apiKeySvc.UpdateKeyLastUsed(keyCtx.KeyID)
			}

			return err
		}
	}
}
