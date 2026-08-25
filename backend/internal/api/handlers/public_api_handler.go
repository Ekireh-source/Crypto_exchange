package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"exchange/internal/api/middleware"
	"exchange/internal/models"
	"exchange/internal/services"
)

// PublicAPIHandler serves the developer-facing API endpoints.
// These are separate from the UI wallet handlers and use API key auth.
type PublicAPIHandler struct {
	walletSvc *services.WalletService
}

// NewPublicAPIHandler creates a new PublicAPIHandler.
func NewPublicAPIHandler(walletSvc *services.WalletService) *PublicAPIHandler {
	return &PublicAPIHandler{walletSvc: walletSvc}
}

// GetAssets returns the list of supported assets with live prices.
// Accessible with both pk_ and sk_ keys.
//
// @Summary List supported assets
// @Description Returns a list of all supported crypto and fiat assets along with their current live prices.
// @Tags Assets
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{} "List of assets"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /assets [get]
func (h *PublicAPIHandler) GetAssets(c echo.Context) error {
	assets, err := h.walletSvc.GetAssets(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch assets",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"object": "list",
		"data":   assets,
	})
}

// GetPortfolio returns the developer's wallet portfolio (balances + USD values).
// Accessible with both pk_ and sk_ keys.
//
// @Summary Get wallet portfolio
// @Description Retrieves the developer's wallet portfolio including asset balances and their total USD value.
// @Tags Wallet
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{} "Portfolio data"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /wallet/portfolio [get]
func (h *PublicAPIHandler) GetPortfolio(c echo.Context) error {
	userID := middleware.GetUserID(c)

	portfolio, totalUSD, err := h.walletSvc.GetPortfolio(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch portfolio",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"object":          "portfolio",
		"total_usd_value": totalUSD,
		"assets":          portfolio,
	})
}

// GetDepositAddress generates or retrieves a deposit address for a specific asset.
// Accessible with both pk_ and sk_ keys.
//
// @Summary Get deposit address
// @Description Generates or retrieves a permanent deposit address for a specific asset ID.
// @Tags Wallet
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param assetID path int true "Asset ID"
// @Success 200 {object} map[string]interface{} "Deposit Address"
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /wallet/deposit/{assetID} [get]
func (h *PublicAPIHandler) GetDepositAddress(c echo.Context) error {
	userID := middleware.GetUserID(c)
	assetIDStr := c.Param("assetID")

	assetID, err := strconv.Atoi(assetIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]string{
			"error": "Invalid asset_id parameter",
		})
	}

	da, err := h.walletSvc.GetDepositAddress(c.Request().Context(), userID, assetID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"object":  "deposit_address",
		"data":    da,
	})
}

// SendCrypto processes an outbound crypto transfer.
// Requires a SECRET key (sk_). Publishable keys will be rejected by middleware.
//
// @Summary Send crypto
// @Description Processes an outbound crypto transfer to an external address. Requires a secret key (sk_).
// @Tags Wallet
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body services.SendRequest true "Send request payload"
// @Success 200 {object} map[string]interface{} "Transaction details"
// @Failure 400 {object} map[string]string "Bad Request"
// @Router /wallet/send [post]
func (h *PublicAPIHandler) SendCrypto(c echo.Context) error {
	userID := middleware.GetUserID(c)

	var req services.SendRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if req.AssetID <= 0 || req.ToAddress == "" || req.Amount == "" {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]string{
			"error": "asset_id, to_address, and amount are required",
		})
	}

	tx, err := h.walletSvc.SendCrypto(c.Request().Context(), userID, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"object": "transaction",
		"data":   tx,
	})
}

// SwapCrypto executes an internal asset swap.
// Requires a SECRET key (sk_). Publishable keys will be rejected by middleware.
//
// @Summary Swap crypto
// @Description Executes an internal swap between two supported assets based on live prices. Requires a secret key (sk_).
// @Tags Wallet
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.SwapRequest true "Swap request payload"
// @Success 200 {object} map[string]interface{} "Swap details"
// @Failure 400 {object} map[string]string "Bad Request"
// @Router /wallet/swap [post]
func (h *PublicAPIHandler) SwapCrypto(c echo.Context) error {
	userID := middleware.GetUserID(c)

	var req models.SwapRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	swap, err := h.walletSvc.SwapCrypto(c.Request().Context(), userID, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"object": "swap",
		"data":   swap,
	})
}

// GetTransactions returns paginated transactions for the developer's account.
// Accessible with both pk_ and sk_ keys.
//
// @Summary List transactions
// @Description Retrieves paginated transaction history for the developer's account.
// @Tags Wallet
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param limit query int false "Number of records to return"
// @Param page query int false "Page number"
// @Param type query string false "Filter by transaction type (e.g. deposit, withdrawal, swap)"
// @Success 200 {object} map[string]interface{} "List of transactions"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /wallet/transactions [get]
func (h *PublicAPIHandler) GetTransactions(c echo.Context) error {
	userID := middleware.GetUserID(c)

	limitStr := c.QueryParam("limit")
	pageStr := c.QueryParam("page")
	txType := c.QueryParam("type")

	limit, _ := strconv.Atoi(limitStr)
	page, _ := strconv.Atoi(pageStr)

	txs, total, err := h.walletSvc.GetTransactions(c.Request().Context(), userID, limit, page, txType)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch transactions",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"object":       "list",
		"data":         txs,
		"total":        total,
		"page":         page,
		"limit":        limit,
	})
}
