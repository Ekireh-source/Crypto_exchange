package api

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"context"
	"exchange/config"
	_ "exchange/docs" // swagger docs
	"exchange/internal/api/handlers"
	"exchange/internal/api/middleware"
	"github.com/swaggo/echo-swagger"
	"exchange/internal/blockchain"
	"exchange/internal/blockchain/bsc"
	"exchange/internal/blockchain/tron"
	"exchange/internal/db"
	"exchange/internal/db/repository"
	"exchange/internal/models"
	"exchange/internal/services"
	"log"
	"time"
)

// RegisterRoutes wires up all HTTP endpoints.
func RegisterRoutes(e *echo.Echo, cfg *config.Config, pool *db.Pool, adapters map[models.Network]blockchain.BlockchainAdapter, scanners map[models.Network]blockchain.DepositScanner) {
	// ── Repositories ──────────────────────────────────────────────────────────
	userRepo := repository.NewUserRepo(pool)
	walletRepo := repository.NewWalletRepo(pool)
	p2pRepo := repository.NewP2PRepo(pool)
	idempotencyRepo := repository.NewIdempotencyRepo(pool)

	// ── Services ──────────────────────────────────────────────────────────────
	emailSvc, err := services.NewEmailService(cfg)
	if err != nil {
		log.Printf("Warning: Failed to initialize email service (emails will not be sent): %v", err)
	}

	authSvc := services.NewAuthService(cfg, userRepo, emailSvc)
	
	wsHub := services.NewWSHub()
	go wsHub.Run()
	
	priceSvc := services.NewPriceService(cfg.CoinGeckoAPIKey, wsHub)
	walletSvc := services.NewWalletService(cfg, walletRepo, priceSvc, adapters, scanners)
	sweeperSvc := services.NewSweeperService(cfg, walletRepo, priceSvc, adapters)
	withdrawalMonitor := services.NewWithdrawalMonitor(cfg, walletRepo, walletSvc, adapters)
	p2pSvc := services.NewP2PService(p2pRepo)
	adminHandler := handlers.NewAdminHandler(userRepo, walletRepo)

	// ── Background Scanners ───────────────────────────────────────────────────
	ctx := context.Background()

	// 0. Start Sweeper Service and Withdrawal Monitor
	sweeperSvc.Start(ctx)
	withdrawalMonitor.Start(ctx)
	go priceSvc.StartPriceTicker(ctx)

	// 1. Initialize BSC Monitor
	if bscAdapter, ok := adapters[models.NetworkBSC].(*bsc.Client); ok {
		tokenContracts := []string{}
		if cfg.USDTBSCContract != "" {
			tokenContracts = append(tokenContracts, cfg.USDTBSCContract)
		}
		monitor := bsc.NewMonitor(bscAdapter, tokenContracts, func(tx blockchain.IncomingTx) {
			err := walletSvc.ProcessDeposit(context.Background(), tx)
			if err != nil {
				log.Printf("ERROR processing BSC deposit: %v", err)
			} else {
				log.Printf("Successfully processed BSC deposit for tx: %s", tx.TxHash)
			}
		}, 3*time.Second, cfg.BSCMinConfirmations)
		scanners[models.NetworkBSC] = monitor
		go monitor.Start(ctx)
	} else {
		log.Printf("ERROR: adapters[models.NetworkBSC] is not *bsc.Client, it is %T", adapters[models.NetworkBSC])
	}

	// 2. Initialize TRON Monitor
	if tronAdapter, ok := adapters[models.NetworkTRON].(*tron.Client); ok {
		tokenContracts := []string{}
		if cfg.USDTTRC20Contract != "" {
			tokenContracts = append(tokenContracts, cfg.USDTTRC20Contract)
		}
		monitor := tron.NewTronMonitor(tronAdapter, tokenContracts, func(tx blockchain.IncomingTx) {
			err := walletSvc.ProcessDeposit(context.Background(), tx)
			if err != nil {
				log.Printf("ERROR processing TRON deposit: %v", err)
			} else {
				log.Printf("Successfully processed TRON deposit for tx: %s", tx.TxHash)
			}
		}, 3*time.Second, cfg.TronMinConfirmations)
		scanners[models.NetworkTRON] = monitor
		go monitor.Start(ctx)
	} else {
		log.Printf("ERROR: adapters[models.NetworkTRON] is not *tron.Client, it is %T", adapters[models.NetworkTRON])
	}

	// 3. Load all existing deposit addresses from DB and register them with their respective scanners
	assets, err := walletRepo.GetAssets(ctx)
	if err == nil {
		assetNetworkMap := make(map[int]models.Network)
		for _, asset := range assets {
			assetNetworkMap[asset.ID] = asset.Network
		}

		das, err := walletRepo.GetAllDepositAddresses(ctx)
		if err == nil {
			for _, da := range das {
				network := assetNetworkMap[da.AssetID]
				if scanner, ok := scanners[network]; ok {
					scanner.AddAddress(da.Address)
				}
			}
		} else {
			log.Printf("ERROR loading deposit addresses: %v", err)
		}
	} else {
		log.Printf("ERROR loading assets: %v", err)
	}

	// ── Handlers ──────────────────────────────────────────────────────────────
	authHandler := handlers.NewAuthHandler(authSvc)
	walletHandler := handlers.NewWalletHandler(walletSvc)
	p2pHandler := handlers.NewP2PHandler(p2pSvc)
	wsHandler := handlers.NewWSHandler(wsHub)

	// ── Developer API Setup ───────────────────────────────────────────────────
	apiKeyRepo := repository.NewAPIKeyRepo(pool)
	apiKeySvc := services.NewAPIKeyService(apiKeyRepo)
	webhookSvc := services.NewWebhookService(apiKeyRepo)
	devHandler := handlers.NewDeveloperHandler(apiKeySvc, webhookSvc)
	publicAPIHandler := handlers.NewPublicAPIHandler(walletSvc)

	v1 := e.Group("/v1")

	// ── Public Routes ─────────────────────────────────────────────────────────
	auth := v1.Group("/auth")
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)
	auth.POST("/refresh", authHandler.Refresh)
	auth.POST("/logout", authHandler.Logout)
	auth.GET("/verify-email", authHandler.VerifyEmail)

	v1.GET("/ws/prices", wsHandler.ServeWS)

	// ── Protected Routes (JWT — your UI) ──────────────────────────────────────
	protected := v1.Group("")
	protected.Use(middleware.RequireAuth(cfg.JWTSecret))

	idempotentMW := middleware.RequireIdempotency(idempotencyRepo)
	verifiedMW := middleware.RequireEmailVerified()

	wallet := protected.Group("/wallet")
	wallet.GET("/assets", walletHandler.GetAssets)
	wallet.GET("/portfolio", walletHandler.GetPortfolio)
	wallet.GET("/deposit/:assetID", walletHandler.GetDepositAddress)
	wallet.POST("/send", walletHandler.SendCrypto, verifiedMW, idempotentMW)
	wallet.POST("/swap", walletHandler.HandleSwap, verifiedMW, idempotentMW)
	wallet.GET("/transactions", walletHandler.GetTransactions)
	wallet.GET("/transactions/:id", walletHandler.GetTransaction)
	wallet.GET("/watchlist", walletHandler.GetWatchlist)
	wallet.POST("/watchlist/toggle", walletHandler.ToggleWatchlist)

	user := protected.Group("/user")
	user.GET("/profile", authHandler.GetProfile)

	p2p := protected.Group("/p2p")
	p2p.POST("/orders", p2pHandler.CreateOrder, verifiedMW, idempotentMW)
	p2p.GET("/orders", p2pHandler.ListActiveOrders)
	p2p.POST("/orders/:id/trade", p2pHandler.CreateTrade, verifiedMW, idempotentMW)
	p2p.GET("/trades", p2pHandler.ListUserTrades)
	p2p.GET("/trades/:id", p2pHandler.GetTrade)
	p2p.POST("/trades/:id/pay", p2pHandler.MarkTradePaid, verifiedMW, idempotentMW)
	p2p.POST("/trades/:id/release", p2pHandler.ReleaseTrade, verifiedMW, idempotentMW)
	p2p.POST("/trades/:id/cancel", p2pHandler.CancelTrade, verifiedMW, idempotentMW)

	// ── Developer Portal Routes (JWT — for managing apps/keys) ────────────────
	developer := protected.Group("/developer")
	developer.POST("/apps", devHandler.CreateApp)
	developer.GET("/apps", devHandler.ListApps)
	developer.DELETE("/apps/:appID", devHandler.DeleteApp)
	developer.PUT("/apps/:appID/status", devHandler.UpdateAppStatus)
	developer.GET("/apps/:appID/keys", devHandler.ListKeys)
	developer.POST("/apps/:appID/keys/regenerate", devHandler.RegenerateKeys)
	developer.DELETE("/apps/:appID/keys/:keyID", devHandler.RevokeKey)
	developer.GET("/apps/:appID/logs", devHandler.GetLogs)
	developer.POST("/apps/:appID/webhooks", devHandler.CreateWebhook)
	developer.GET("/apps/:appID/webhooks", devHandler.ListWebhooks)
	developer.DELETE("/apps/:appID/webhooks/:webhookID", devHandler.DeleteWebhook)

	// ── Admin Routes (Protected, Admin/Superadmin only) ───────────────────────
	admin := protected.Group("/admin")
	admin.Use(middleware.RequireRole("admin", "superadmin"))
	admin.GET("/users", adminHandler.GetUsers)
	admin.PUT("/users/:id/role", adminHandler.UpdateUserRole, middleware.RequireRole("superadmin"))
	admin.GET("/transactions", adminHandler.GetTransactions)


	// ── Swagger Documentation ─────────────────────────────────────────────────
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// ── Public Developer API (API Key auth — for third-party consumers) ───────
	publicAPI := e.Group("/api/v1")
	publicAPI.Use(middleware.RequireAPIKey(apiKeySvc))
	publicAPI.Use(middleware.RateLimitByAPIKey())
	publicAPI.Use(middleware.APIRequestLogger(apiKeySvc))

	publicAPI.GET("/assets", publicAPIHandler.GetAssets)
	publicAPI.GET("/wallet/portfolio", publicAPIHandler.GetPortfolio)
	publicAPI.GET("/wallet/deposit/:assetID", publicAPIHandler.GetDepositAddress)
	publicAPI.POST("/wallet/send", publicAPIHandler.SendCrypto, middleware.RequireSignature())
	publicAPI.POST("/wallet/swap", publicAPIHandler.SwapCrypto, middleware.RequireSignature())
	publicAPI.GET("/wallet/transactions", publicAPIHandler.GetTransactions)
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
