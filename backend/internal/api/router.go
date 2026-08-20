package api

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"exchange/config"
	"exchange/internal/api/handlers"
	"exchange/internal/api/middleware"
	"exchange/internal/blockchain"
	"exchange/internal/blockchain/bsc"
	"exchange/internal/blockchain/tron"
	"exchange/internal/db"
	"exchange/internal/db/repository"
	"exchange/internal/models"
	"exchange/internal/services"
	"context"
	"log"
	"time"
)

// RegisterRoutes wires up all HTTP endpoints.
func RegisterRoutes(e *echo.Echo, cfg *config.Config, pool *db.Pool, adapters map[models.Network]blockchain.BlockchainAdapter, scanners map[models.Network]blockchain.DepositScanner) {
	// ── Repositories ──────────────────────────────────────────────────────────
	userRepo := repository.NewUserRepo(pool)
	walletRepo := repository.NewWalletRepo(pool)

	// ── Services ──────────────────────────────────────────────────────────────
	authSvc := services.NewAuthService(cfg, userRepo)
	priceSvc := services.NewPriceService(cfg.CoinGeckoAPIKey)
	walletSvc := services.NewWalletService(cfg, walletRepo, priceSvc, adapters, scanners)

	// ── Background Scanners ───────────────────────────────────────────────────
	ctx := context.Background()

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
		}, 3*time.Second)
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
		}, 3*time.Second)
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
	wallet.GET("/assets", walletHandler.GetAssets)
	wallet.GET("/portfolio", walletHandler.GetPortfolio)
	wallet.GET("/deposit/:assetID", walletHandler.GetDepositAddress)
	wallet.POST("/send", walletHandler.SendCrypto)
	wallet.GET("/transactions", walletHandler.GetTransactions)
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
