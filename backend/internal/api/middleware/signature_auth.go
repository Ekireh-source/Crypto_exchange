package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"exchange/internal/models"
)

// RequireSignature enforces HMAC-SHA256 signatures on API requests.
// It expects RequireAPIKey to have already executed and injected rawAPIKey.
func RequireSignature() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			keyCtx, ok := c.Get("apiKeyCtx").(*models.APIKeyContext)
			if !ok || keyCtx == nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing API key context")
			}

			// Enforce signature only on secret keys
			if keyCtx.KeyType != models.APIKeyTypeSecret {
				return echo.NewHTTPError(http.StatusForbidden, "Only secret keys (sk_) can be used with signed requests")
			}

			timestampStr := c.Request().Header.Get("X-Timestamp")
			clientSignature := c.Request().Header.Get("X-Signature")

			if timestampStr == "" || clientSignature == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing X-Timestamp or X-Signature header")
			}

			// 1. Replay Protection: verify timestamp is within 5 minutes
			timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "Invalid X-Timestamp format")
			}

			now := time.Now().Unix()
			if now-timestamp > 300 || timestamp-now > 300 {
				return echo.NewHTTPError(http.StatusUnauthorized, "Request expired (timestamp outside 5-minute window)")
			}

			// 2. Read request body safely
			var bodyBytes []byte
			if c.Request().Body != nil {
				bodyBytes, _ = io.ReadAll(c.Request().Body)
				c.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // restore for next handlers
			}

			// 3. Recompute expected signature
			rawAPIKey, ok := c.Get("rawAPIKey").(string)
			if !ok || rawAPIKey == "" {
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve raw API key for signature verification")
			}

			method := c.Request().Method
			path := c.Request().URL.Path
			if c.Request().URL.RawQuery != "" {
				path += "?" + c.Request().URL.RawQuery
			}

			// Payload to sign: timestamp + method + path + body
			payload := timestampStr + method + path + string(bodyBytes)

			mac := hmac.New(sha256.New, []byte(rawAPIKey))
			mac.Write([]byte(payload))
			expectedSignature := hex.EncodeToString(mac.Sum(nil))

			// 4. Constant-time string comparison to prevent timing attacks
			if subtle.ConstantTimeCompare([]byte(clientSignature), []byte(expectedSignature)) != 1 {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid request signature")
			}

			return next(c)
		}
	}
}
