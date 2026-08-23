package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"exchange/internal/api/middleware"
	"exchange/internal/models"
	"exchange/internal/services"
)

type P2PHandler struct {
	p2pSvc *services.P2PService
}

func NewP2PHandler(p2pSvc *services.P2PService) *P2PHandler {
	return &P2PHandler{p2pSvc: p2pSvc}
}

type CreateOrderRequest struct {
	AssetID       int    `json:"asset_id"`
	FiatCurrency  string `json:"fiat_currency"`
	Rate          string `json:"rate"`
	MinAmount     string `json:"min_amount"`
	MaxAmount     string `json:"max_amount"`
	TotalAmount   string `json:"total_amount"`
	PaymentMethod string `json:"payment_method"`
}

func (h *P2PHandler) CreateOrder(c echo.Context) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	var req CreateOrderRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	order, err := h.p2pSvc.CreateOrder(
		c.Request().Context(),
		userID,
		req.AssetID,
		req.FiatCurrency,
		req.Rate,
		req.MinAmount,
		req.MaxAmount,
		req.TotalAmount,
		req.PaymentMethod,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, order)
}

func (h *P2PHandler) ListActiveOrders(c echo.Context) error {
	orders, err := h.p2pSvc.ListActiveOrders(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if orders == nil {
		orders = []*models.P2POrder{}
	}
	return c.JSON(http.StatusOK, orders)
}

type CreateTradeRequest struct {
	Amount     string `json:"amount"`
	FiatAmount string `json:"fiat_amount"`
}

func (h *P2PHandler) CreateTrade(c echo.Context) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	orderIDStr := c.Param("id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid order ID")
	}

	var req CreateTradeRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	trade, err := h.p2pSvc.CreateTrade(c.Request().Context(), userID, orderID, req.Amount, req.FiatAmount)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, trade)
}

func (h *P2PHandler) ListUserTrades(c echo.Context) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	trades, err := h.p2pSvc.ListUserTrades(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if trades == nil {
		trades = []*models.P2PTrade{}
	}
	return c.JSON(http.StatusOK, trades)
}

func (h *P2PHandler) GetTrade(c echo.Context) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	tradeIDStr := c.Param("id")
	tradeID, err := uuid.Parse(tradeIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid trade ID")
	}

	trade, err := h.p2pSvc.GetTrade(c.Request().Context(), tradeID, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, trade)
}

func (h *P2PHandler) MarkTradePaid(c echo.Context) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	tradeIDStr := c.Param("id")
	tradeID, err := uuid.Parse(tradeIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid trade ID")
	}

	if err := h.p2pSvc.MarkTradePaid(c.Request().Context(), tradeID, userID); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Trade marked as paid"})
}

func (h *P2PHandler) ReleaseTrade(c echo.Context) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	tradeIDStr := c.Param("id")
	tradeID, err := uuid.Parse(tradeIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid trade ID")
	}

	if err := h.p2pSvc.ReleaseTrade(c.Request().Context(), tradeID, userID); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Trade released successfully"})
}

func (h *P2PHandler) CancelTrade(c echo.Context) error {
	userID := middleware.GetUserID(c)
	if userID == uuid.Nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	tradeIDStr := c.Param("id")
	tradeID, err := uuid.Parse(tradeIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid trade ID")
	}

	if err := h.p2pSvc.CancelTrade(c.Request().Context(), tradeID, userID); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Trade cancelled successfully"})
}
