package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"exchange/internal/api/middleware"
	"exchange/internal/services"
)

type AuthHandler struct {
	authSvc *services.AuthService
}

func NewAuthHandler(authSvc *services.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

func setTokenCookies(c echo.Context, tokens *services.TokenPair) {
	secure := c.Scheme() == "https"
	
	c.SetCookie(&http.Cookie{
		Name:     "access_token",
		Value:    tokens.AccessToken,
		Path:     "/",
		Expires:  time.Now().Add(15 * time.Minute),
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
	
	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    tokens.RefreshToken,
		Path:     "/",
		Expires:  time.Now().Add(168 * time.Hour),
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

type RegisterRequest struct {
	Email           string `json:"email" form:"email"`
	Password        string `json:"password" form:"password"`
	ConfirmPassword string `json:"confirm_password" form:"confirm_password"`
	ReferralCode    string `json:"referral_code" form:"referral_code"`
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if req.Email == "" || len(req.Password) < 8 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid email or password too short")
	}

	if req.Password != req.ConfirmPassword {
		return echo.NewHTTPError(http.StatusBadRequest, "Passwords do not match")
	}


	user, err := h.authSvc.Register(c.Request().Context(), req.Email, req.Password, req.ReferralCode)
	if err != nil {
		if errors.Is(err, services.ErrUserExists) {
			return echo.NewHTTPError(http.StatusConflict, "User already exists")
		}
		if errors.Is(err, services.ErrInvalidReferral) {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid referral code")
		}
		return err
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "Registration successful",
		"user": map[string]interface{}{
			"id":            user.ID,
			"email":         user.Email,
			"referral_code": user.ReferralCode,
			"role":          user.Role,
		},
	})
}

type LoginRequest struct {
	Email    string `json:"email" form:"email"`
	Password string `json:"password" form:"password"`
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	user, tokens, err := h.authSvc.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid email or password")
		}
		return err
	}

	setTokenCookies(c, tokens)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user": map[string]interface{}{
			"id":            user.ID,
			"email":         user.Email,
			"referral_code": user.ReferralCode,
			"role":          user.Role,
		},
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	})
}

func (h *AuthHandler) VerifyEmail(c echo.Context) error {
	token := c.QueryParam("token")
	if token == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing token")
	}

	err := h.authSvc.VerifyEmail(c.Request().Context(), token)
	if err != nil {
		if errors.Is(err, services.ErrInvalidToken) || err.Error() == "verification token expired" {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid or expired token")
		}
		return err
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Email successfully verified",
	})
}


type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" form:"refresh_token"`
}

func (h *AuthHandler) Refresh(c echo.Context) error {
	var req RefreshRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if req.RefreshToken == "" {
		if cookie, err := c.Cookie("refresh_token"); err == nil {
			req.RefreshToken = cookie.Value
		}
	}

	tokens, err := h.authSvc.RefreshToken(c.Request().Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, services.ErrInvalidToken) {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired refresh token")
		}
		return err
	}

	setTokenCookies(c, tokens)

	return c.JSON(http.StatusOK, tokens)
}

func (h *AuthHandler) Logout(c echo.Context) error {
	secure := c.Scheme() == "https"
	
	c.SetCookie(&http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
	
	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
	
	return c.JSON(http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

func (h *AuthHandler) GetProfile(c echo.Context) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	user, stats, referredUsers, err := h.authSvc.GetProfile(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	// Format referred users
	formattedReferred := make([]map[string]interface{}, 0, len(referredUsers))
	for _, ru := range referredUsers {
		formattedReferred = append(formattedReferred, map[string]interface{}{
			"id":         ru.ID,
			"email":      ru.Email,
			"phone":      ru.Phone,
			"kyc_status": ru.KYCStatus,
			"created_at": ru.CreatedAt,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user": map[string]interface{}{
			"id":            user.ID,
			"email":         user.Email,
			"phone":         user.Phone,
			"kyc_status":    user.KYCStatus,
			"role":          user.Role,
			"created_at":    user.CreatedAt,
			"referral_code": user.ReferralCode,
		},
		"referral_stats": stats,
		"referred_users": formattedReferred,
	})
}
