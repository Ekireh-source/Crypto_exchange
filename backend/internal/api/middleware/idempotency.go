package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"exchange/internal/db/repository"
	"github.com/labstack/echo/v4"
)

type bodyDumpResponseWriter struct {
	echo.Response
	body *bytes.Buffer
}

func (w *bodyDumpResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.Response.Write(b)
}

// RequireIdempotency enforces an idempotency key for the endpoint.
func RequireIdempotency(repo *repository.IdempotencyRepo) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Extract Idempotency Key
			key := c.Request().Header.Get("Idempotency-Key")
			if key == "" {
				return c.JSON(http.StatusBadRequest, map[string]string{
					"error": "Idempotency-Key header is required",
				})
			}

			// Extract User ID from context (assuming RequireAuth middleware has run)
			userIDRaw := c.Get("userID")
			if userIDRaw == nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Unauthorized",
				})
			}
			userID := ""
			switch v := userIDRaw.(type) {
			case string:
				userID = v
			case interface{ String() string }:
				userID = v.String()
			default:
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Invalid user ID format in context",
				})
			}

			// Set expiration for 24 hours
			expiresAt := time.Now().Add(24 * time.Hour)

			// Try to acquire the lock
			idempKey, err := repo.AcquireLock(c.Request().Context(), userID, key, c.Request().Method, c.Request().URL.Path, expiresAt)
			if err != nil {
				if err == repository.ErrKeyExists {
					// Key exists
					switch idempKey.Status {
					case repository.StatusInProgress:
						return c.JSON(http.StatusConflict, map[string]string{
							"error": "A request with this Idempotency-Key is already in progress.",
						})
					case repository.StatusCompleted:
						// Return cached response
						var body interface{}
						if len(idempKey.ResponseBody) > 0 {
							_ = json.Unmarshal(idempKey.ResponseBody, &body)
						}
						return c.JSON(idempKey.ResponseStatus, body)
					}
				}
				// Database error
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process idempotency key")
			}

			// Setup response wrapper to capture the body and status code
			res := &bodyDumpResponseWriter{
				Response: *c.Response(),
				body:     new(bytes.Buffer),
			}
			c.Response().Writer = res

			// Proceed with the request
			handlerErr := next(c)

			// Determine final status code. Note: echo handles errors, so if handlerErr != nil,
			// the actual status code might not be set in the writer yet. We will extract it.
			statusCode := res.Status
			var responseBody []byte

			if handlerErr != nil {
				// If error was returned, Echo's central error handler usually formats it.
				// We can try to extract the status from HTTPError.
				if he, ok := handlerErr.(*echo.HTTPError); ok {
					statusCode = he.Code
					responseBody, _ = json.Marshal(map[string]interface{}{"error": he.Message})
				} else {
					statusCode = http.StatusInternalServerError
					responseBody, _ = json.Marshal(map[string]interface{}{"error": "Internal Server Error"})
				}
			} else {
				responseBody = res.body.Bytes()
			}

			// If it's a 5xx error, we might still save it, or we could delete the key to allow retries.
			// Standard practice is to save it so the client knows it failed uniformly.
			_ = repo.SaveResponse(c.Request().Context(), userID, key, statusCode, responseBody)

			return handlerErr
		}
	}
}
