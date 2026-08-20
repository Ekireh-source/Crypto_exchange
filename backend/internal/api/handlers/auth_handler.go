package handlers

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"exchange/internal/services"
)

type AuthHandler struct {
	authSvc *services.AuthService
}

func NewAuthHandler(authSvc *services.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

type RegisterRequest struct {
	Email        string `json:"email" form:"email"`
	Password     string `json:"password" form:"password"`
	ReferralCode string `json:"referral_code" form:"referral_code"`
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if req.Email == "" || len(req.Password) < 8 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid email or password too short")
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

	tokens, err := h.authSvc.GenerateTokenPair(user.ID)
	if err != nil {
		return err
	}

	// Omit password hash in response
	return c.JSON(http.StatusCreated, map[string]interface{}{
		"user": map[string]interface{}{
			"id":            user.ID,
			"email":         user.Email,
			"referral_code": user.ReferralCode,
		},
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
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

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user": map[string]interface{}{
			"id":            user.ID,
			"email":         user.Email,
			"referral_code": user.ReferralCode,
		},
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
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

	tokens, err := h.authSvc.RefreshToken(c.Request().Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, services.ErrInvalidToken) {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired refresh token")
		}
		return err
	}

	return c.JSON(http.StatusOK, tokens)
}
