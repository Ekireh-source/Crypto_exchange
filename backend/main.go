package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"exchange/config"
	"exchange/internal/api"
	"exchange/internal/blockchain"
	"exchange/internal/blockchain/bsc"
	"exchange/internal/blockchain/tron"
	"exchange/internal/db"
	"exchange/internal/models"
)

func main() {
	// ── Load configuration ────────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}

	// ── Context with graceful shutdown ────────────────────────────────────────
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// ── Database ──────────────────────────────────────────────────────────────
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connecting to database: %v", err)
	}
	defer pool.Close()
	log.Println("✓ database connected")

	// ── Migrations ────────────────────────────────────────────────────────────
	migrationsDir := "internal/db/migrations"
	if err := db.Migrate(ctx, pool, migrationsDir); err != nil {
		log.Fatalf("running migrations: %v", err)
	}
	log.Println("✓ migrations applied")

	// ── Echo HTTP server ──────────────────────────────────────────────────────
	e := echo.New()
	e.HideBanner = true
	e.HTTPErrorHandler = api.ErrorHandler

	// Global middleware
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Use(middleware.CORS())
	e.Use(middleware.RequestID())

	// ── Blockchain Adapters ───────────────────────────────────────────────────
	bscClient, err := bsc.NewClient(ctx, cfg.ActiveBSCRPC(), "BSC")
	if err != nil {
		log.Fatalf("connecting to BSC: %v", err)
	}
	defer bscClient.Close()

	tronClient := tron.NewClient(cfg.ActiveTronBaseURL(), cfg.TronGridAPIKey, "TRON")

	adapters := map[models.Network]blockchain.BlockchainAdapter{
		models.NetworkBSC:  bscClient,
		models.NetworkTRON: tronClient,
	}

	// Register all application routes
	api.RegisterRoutes(e, cfg, pool, adapters)

	// ── Start server ──────────────────────────────────────────────────────────
	go func() {
		addr := ":" + cfg.Port
		log.Printf("✓ server listening on %s", addr)
		if err := e.Start(addr); err != nil {
			log.Printf("server stopped: %v", err)
		}
	}()

	// ── Wait for shutdown signal ──────────────────────────────────────────────
	<-ctx.Done()
	log.Println("shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := e.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown error: %v", err)
	}
	log.Println("goodbye")
}
