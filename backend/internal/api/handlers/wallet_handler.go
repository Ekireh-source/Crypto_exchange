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

// GetAssets returns a list of supported assets.
func (h *WalletHandler) GetAssets(c echo.Context) error {
	assets, err := h.walletSvc.GetAssets(c.Request().Context())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, assets)
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

// SendCrypto processes a request to send crypto to a recipient address.
func (h *WalletHandler) SendCrypto(c echo.Context) error {
	userID := middleware.GetUserID(c)

	var req services.SendRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if req.AssetID <= 0 || req.ToAddress == "" || req.Amount == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "asset_id, to_address, and amount are required")
	}

	tx, err := h.walletSvc.SendCrypto(c.Request().Context(), userID, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, tx)
}

// GetTransactions returns a paginated list of transactions for the authenticated user.
func (h *WalletHandler) GetTransactions(c echo.Context) error {
	userID := middleware.GetUserID(c)

	limitStr := c.QueryParam("limit")
	pageStr := c.QueryParam("page")
	txType := c.QueryParam("type")

	limit, _ := strconv.Atoi(limitStr)
	page, _ := strconv.Atoi(pageStr)

	txs, total, err := h.walletSvc.GetTransactions(c.Request().Context(), userID, limit, page, txType)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"transactions": txs,
		"total":        total,
		"page":         page,
		"limit":        limit,
	})
}


