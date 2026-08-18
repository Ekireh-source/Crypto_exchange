package api

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"exchange/config"
	"exchange/internal/api/handlers"
	"exchange/internal/api/middleware"
	"exchange/internal/blockchain"
	"exchange/internal/db"
	"exchange/internal/db/repository"
	"exchange/internal/models"
	"exchange/internal/services"
)

// RegisterRoutes wires up all HTTP endpoints.
func RegisterRoutes(e *echo.Echo, cfg *config.Config, pool *db.Pool, adapters map[models.Network]blockchain.BlockchainAdapter) {
	// ── Repositories ──────────────────────────────────────────────────────────
	userRepo := repository.NewUserRepo(pool)
	walletRepo := repository.NewWalletRepo(pool)

	// ── Services ──────────────────────────────────────────────────────────────
	authSvc := services.NewAuthService(cfg, userRepo)
	priceSvc := services.NewPriceService(cfg.CoinGeckoAPIKey)
	walletSvc := services.NewWalletService(cfg, walletRepo, priceSvc, adapters)

	// ── Handlers ──────────────────────────────────────────────────────────────
	authHandler := handlers.NewAuthHandler(authSvc)
	walletHandler := handlers.NewWalletHandler(walletSvc)

	v1 := e.Group("/v1")

	// ── Public Routes ─────────────────────────────────────────────────────────
	auth := v1.Group("/auth")
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)
	auth.POST("/refresh", authHandler.Refresh)

	// ── Protected Routes ──────────────────────────────────────────────────────
	protected := v1.Group("")
	protected.Use(middleware.RequireAuth(cfg.JWTSecret))
	
	wallet := protected.Group("/wallet")
	wallet.GET("/portfolio", walletHandler.GetPortfolio)
	wallet.GET("/deposit/:assetID", walletHandler.GetDepositAddress)
}

// ErrorHandler formats HTTP errors as JSON instead of plaintext.
func ErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	code := http.StatusInternalServerError
	message := "Internal Server Error"

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		message = fmt.Sprintf("%v", he.Message)
	}

	c.JSON(code, map[string]interface{}{
		"error": message,
	})
}
