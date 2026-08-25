package handlers

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"exchange/internal/db/repository"
)

type AdminHandler struct {
	userRepo   *repository.UserRepo
	walletRepo *repository.WalletRepo
}

func NewAdminHandler(userRepo *repository.UserRepo, walletRepo *repository.WalletRepo) *AdminHandler {
	return &AdminHandler{
		userRepo:   userRepo,
		walletRepo: walletRepo,
	}
}

// GetUsers returns a list of all users.
func (h *AdminHandler) GetUsers(c echo.Context) error {
	users, err := h.userRepo.ListUsers(c.Request().Context())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"users": users,
	})
}

type UpdateRoleRequest struct {
	Role string `json:"role"`
}

// UpdateUserRole updates a user's role (superadmin only).
func (h *AdminHandler) UpdateUserRole(c echo.Context) error {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID")
	}

	var req UpdateRoleRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if req.Role != "user" && req.Role != "admin" && req.Role != "superadmin" {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid role")
	}

	if err := h.userRepo.UpdateRole(c.Request().Context(), userID, req.Role); err != nil {
		return fmt.Errorf("updating user role: %w", err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "User role updated successfully",
	})
}

// GetTransactions returns a list of all transactions across the platform.
func (h *AdminHandler) GetTransactions(c echo.Context) error {
	txs, total, err := h.walletRepo.GetAllTransactions(c.Request().Context(), 100, 0)
	if err != nil {
		return fmt.Errorf("fetching all transactions: %w", err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"transactions": txs,
		"total":        total,
	})
}

