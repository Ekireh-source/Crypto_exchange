package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// RequireAuth is a middleware that validates the JWT access token and
// injects the user ID into the request context.
func RequireAuth(jwtSecret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			var tokenString string

			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
					tokenString = parts[1]
				}
			}

			// If not in header, try cookie
			if tokenString == "" {
				if cookie, err := c.Cookie("access_token"); err == nil {
					tokenString = cookie.Value
				}
			}

			if tokenString == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing Authorization token")
			}
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired token")
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok || claims["type"] != "access" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token type")
			}

			userIDStr, ok := claims["sub"].(string)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token subject")
			}

			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid user ID in token")
			}

			// Inject User ID and Role into context
			c.Set("userID", userID)
			if role, ok := claims["role"].(string); ok {
				c.Set("userRole", role)
			} else {
				c.Set("userRole", "user") // default
			}
			
			if emailVerified, ok := claims["email_verified"].(bool); ok {
				c.Set("emailVerified", emailVerified)
			} else {
				c.Set("emailVerified", false)
			}

			return next(c)
		}
	}
}

// GetUserID is a helper to extract the parsed user ID from context.
func GetUserID(c echo.Context) uuid.UUID {
	id, _ := c.Get("userID").(uuid.UUID)
	return id
}

// GetUserRole extracts the user role from the echo.Context.
func GetUserRole(c echo.Context) string {
	role, ok := c.Get("userRole").(string)
	if !ok {
		return "user"
	}
	return role
}

// IsEmailVerified extracts the verified status from the echo.Context.
func IsEmailVerified(c echo.Context) bool {
	verified, ok := c.Get("emailVerified").(bool)
	if !ok {
		return false
	}
	return verified
}

// RequireRole is a middleware that restricts access to users with one of the specified roles.
// It assumes RequireAuth has already run.
func RequireRole(roles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userRole := GetUserRole(c)
			for _, allowedRole := range roles {
				if userRole == allowedRole {
					return next(c)
				}
			}
			return echo.NewHTTPError(http.StatusForbidden, "Insufficient permissions")
		}
	}
}

// RequireEmailVerified restricts access to users who have verified their email.
func RequireEmailVerified() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !IsEmailVerified(c) {
				return echo.NewHTTPError(http.StatusForbidden, "Email verification required")
			}
			return next(c)
		}
	}
}
