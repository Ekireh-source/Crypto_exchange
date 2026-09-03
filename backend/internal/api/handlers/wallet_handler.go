package handlers

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"exchange/internal/api/middleware"
	"exchange/internal/models"
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

	if txs == nil {
		txs = []models.Transaction{}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"transactions": txs,
		"total":        total,
		"page":         page,
		"limit":        limit,
	})
}

// GetTransaction fetches a single transaction by ID.
func (h *WalletHandler) GetTransaction(c echo.Context) error {
	userID := middleware.GetUserID(c)
	txIDStr := c.Param("id")
	txID, err := uuid.Parse(txIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid transaction id")
	}

	tx, err := h.walletSvc.GetTransactionByID(c.Request().Context(), txID, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if tx == nil {
		return echo.NewHTTPError(http.StatusNotFound, "transaction not found")
	}

	// If it's a swap, we could also attach the swap details, or just return the tx.
	return c.JSON(http.StatusOK, tx)
}

// HandleSwap processes an internal asset swap.
// @Summary Swap assets
// @Description Swap one asset for another internally
// @Tags wallet
// @Accept json
// @Produce json
// @Param request body models.SwapRequest true "Swap Request"
// @Success 201 {object} models.SwapResponse
// @Router /v1/wallet/swap [post]
// @Security BearerAuth
func (h *WalletHandler) HandleSwap(c echo.Context) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	var req models.SwapRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	swap, err := h.walletSvc.SwapCrypto(c.Request().Context(), userID, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, models.SwapResponse{Swap: *swap})
}

func (h *WalletHandler) GetWatchlist(c echo.Context) error {
	userID := middleware.GetUserID(c)
	
	watchlist, err := h.walletSvc.GetWatchlist(c.Request().Context(), userID)
	if err != nil {
		return err
	}
	
	// Handle nil watchlist as empty array
	if watchlist == nil {
		watchlist = []int{}
	}

	return c.JSON(http.StatusOK, watchlist)
}

func (h *WalletHandler) ToggleWatchlist(c echo.Context) error {
	userID := middleware.GetUserID(c)
	
	var req struct {
		AssetID int `json:"asset_id"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if err := h.walletSvc.ToggleWatchlist(c.Request().Context(), userID, req.AssetID); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}
