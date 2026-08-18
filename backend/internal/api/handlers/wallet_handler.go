package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"exchange/internal/api/middleware"
	"exchange/internal/services"
)

type WalletHandler struct {
	walletSvc *services.WalletService
}

func NewWalletHandler(walletSvc *services.WalletService) *WalletHandler {
	return &WalletHandler{walletSvc: walletSvc}
}

// GetPortfolio returns the aggregated balance and USD value for a user.
func (h *WalletHandler) GetPortfolio(c echo.Context) error {
	userID := middleware.GetUserID(c)

	portfolio, totalUSD, err := h.walletSvc.GetPortfolio(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"total_usd_value": totalUSD,
		"assets":          portfolio,
	})
}

// GetDepositAddress fetches or generates a deposit address for a specific asset.
func (h *WalletHandler) GetDepositAddress(c echo.Context) error {
	userID := middleware.GetUserID(c)
	assetIDStr := c.Param("assetID")
	
	assetID, err := strconv.Atoi(assetIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid asset ID")
	}

	da, err := h.walletSvc.GetDepositAddress(c.Request().Context(), userID, assetID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, da)
}
