package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	// Server
	Port           string
	Env            string
	AllowedOrigins []string

	// Auth
	JWTSecret         string
	JWTAccessExpiry   time.Duration
	JWTRefreshExpiry  time.Duration

	// Database
	DatabaseURL string

	// Redis
	RedisURL string

	// HD Wallet
	MasterSeedPhrase string
	EncryptionKey    string // hex-encoded 32-byte key for AES-256

	// BSC / EVM
	BSCMainnetRPC      string
	BSCTestnetRPC      string
	BSCChainID         int64
	BSCTestnetChainID  int64
	USDTBSCContract    string

	// TRON
	TronGridAPIKey     string
	TronGridBaseURL    string
	TronShastaBaseURL  string
	USDTTRC20Contract  string

	// Exchange
	ExchangeFeeRate    float64
	ReferralFeeShare   float64

	// External APIs
	CoinGeckoAPIKey string

	// Hot Wallet / Sweeper
	HotWalletBSCKey         string
	HotWalletBSCAddress     string
	HotWalletTronKey        string
	HotWalletTronAddress    string
	MasterGasWalletBSCKey   string
	MasterGasWalletTronKey  string
	MinSweepUSDThreshold    float64
	WithdrawalFeeUSD     float64
	BSCMinConfirmations  int64
	TronMinConfirmations int64
}

// Load reads .env (if present) then environment variables and returns a Config.
// It returns an error if any required variable is missing.
func Load() (*Config, error) {
	// Load .env file if it exists — ignore error in production (vars set externally)
	_ = godotenv.Load()

	cfg := &Config{}

	// ── Server ────────────────────────────────────────────────────────────────
	cfg.Port = getEnvOrDefault("PORT", "8080")
	cfg.Env = getEnvOrDefault("ENV", "development")
	
	originsStr := getEnvOrDefault("ALLOWED_ORIGINS", "http://localhost:3000")
	var origins []string
	for _, o := range strings.Split(originsStr, ",") {
		origins = append(origins, strings.TrimSpace(o))
	}
	cfg.AllowedOrigins = origins

	// ── Auth ──────────────────────────────────────────────────────────────────
	cfg.JWTSecret = mustGetEnv("JWT_SECRET")

	accessExpiry, err := time.ParseDuration(getEnvOrDefault("JWT_ACCESS_EXPIRY", "15m"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_ACCESS_EXPIRY: %w", err)
	}
	cfg.JWTAccessExpiry = accessExpiry

	refreshExpiry, err := time.ParseDuration(getEnvOrDefault("JWT_REFRESH_EXPIRY", "168h"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_EXPIRY: %w", err)
	}
	cfg.JWTRefreshExpiry = refreshExpiry

	// ── Database ──────────────────────────────────────────────────────────────
	cfg.DatabaseURL = mustGetEnv("DATABASE_URL")

	// ── Redis ─────────────────────────────────────────────────────────────────
	cfg.RedisURL = getEnvOrDefault("REDIS_URL", "redis://localhost:6379")

	// ── HD Wallet ─────────────────────────────────────────────────────────────
	cfg.MasterSeedPhrase = mustGetEnv("MASTER_SEED_PHRASE")
	cfg.EncryptionKey = mustGetEnv("ENCRYPTION_KEY")

	// ── BSC / EVM ─────────────────────────────────────────────────────────────
	cfg.BSCMainnetRPC = getEnvOrDefault("BSC_MAINNET_RPC", "https://bsc-dataseed.binance.org/")
	cfg.BSCTestnetRPC = getEnvOrDefault("BSC_TESTNET_RPC", "https://data-seed-prebsc-1-s1.binance.org:8545/")
	cfg.BSCChainID = int64(mustGetEnvInt("BSC_CHAIN_ID", 56))
	cfg.BSCTestnetChainID = int64(mustGetEnvInt("BSC_TESTNET_CHAIN_ID", 97))
	cfg.USDTBSCContract = getEnvOrDefault("USDT_BSC_CONTRACT", "0x55d398326f99059fF775485246999027B3197955")

	// ── TRON ──────────────────────────────────────────────────────────────────
	cfg.TronGridAPIKey = getEnvOrDefault("TRON_GRID_API_KEY", "")
	cfg.TronGridBaseURL = getEnvOrDefault("TRON_GRID_BASE_URL", "https://api.trongrid.io")
	cfg.TronShastaBaseURL = getEnvOrDefault("TRON_SHASTA_BASE_URL", "https://api.shasta.trongrid.io")
	cfg.USDTTRC20Contract = getEnvOrDefault("USDT_TRC20_CONTRACT", "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t")

	// ── Exchange ──────────────────────────────────────────────────────────────
	feeRate, err := strconv.ParseFloat(getEnvOrDefault("EXCHANGE_FEE_RATE", "0.001"), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid EXCHANGE_FEE_RATE: %w", err)
	}
	cfg.ExchangeFeeRate = feeRate

	referralShare, err := strconv.ParseFloat(getEnvOrDefault("REFERRAL_FEE_SHARE", "0.30"), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid REFERRAL_FEE_SHARE: %w", err)
	}
	cfg.ReferralFeeShare = referralShare

	// ── External APIs ─────────────────────────────────────────────────────────
	cfg.CoinGeckoAPIKey = getEnvOrDefault("COINGECKO_API_KEY", "")

	// ── Hot Wallet / Sweeper ──────────────────────────────────────────────────
	cfg.HotWalletBSCKey = getEnvOrDefault("HOT_WALLET_BSC_KEY", "")
	cfg.HotWalletBSCAddress = getEnvOrDefault("HOT_WALLET_BSC_ADDRESS", "")
	cfg.HotWalletTronKey = getEnvOrDefault("HOT_WALLET_TRON_KEY", "")
	cfg.HotWalletTronAddress = getEnvOrDefault("HOT_WALLET_TRON_ADDRESS", "")
	cfg.MasterGasWalletBSCKey = getEnvOrDefault("MASTER_GAS_WALLET_BSC_KEY", "")
	cfg.MasterGasWalletTronKey = getEnvOrDefault("MASTER_GAS_WALLET_TRON_KEY", "")

	minSweep, _ := strconv.ParseFloat(getEnvOrDefault("MIN_SWEEP_USD_THRESHOLD", "50.0"), 64)
	cfg.MinSweepUSDThreshold = minSweep

	withdrawFee, _ := strconv.ParseFloat(getEnvOrDefault("WITHDRAWAL_FEE_USD", "1.0"), 64)
	cfg.WithdrawalFeeUSD = withdrawFee

	cfg.BSCMinConfirmations = int64(mustGetEnvInt("BSC_MIN_CONFIRMATIONS", 15))
	cfg.TronMinConfirmations = int64(mustGetEnvInt("TRON_MIN_CONFIRMATIONS", 20))

	return cfg, nil
}

// IsDevelopment returns true when running in development mode.
func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}

// IsProduction returns true when running in production mode.
func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

// ActiveBSCRPC returns testnet RPC in development, mainnet in production.
func (c *Config) ActiveBSCRPC() string {
	if c.IsDevelopment() {
		return c.BSCTestnetRPC
	}
	return c.BSCMainnetRPC
}

// ActiveBSCChainID returns the chain ID for the active network.
func (c *Config) ActiveBSCChainID() int64 {
	if c.IsDevelopment() {
		return c.BSCTestnetChainID
	}
	return c.BSCChainID
}

// ActiveTronBaseURL returns Shasta testnet URL in development, mainnet otherwise.
func (c *Config) ActiveTronBaseURL() string {
	if c.IsDevelopment() {
		return c.TronShastaBaseURL
	}
	return c.TronGridBaseURL
}

// ── helpers ───────────────────────────────────────────────────────────────────

func mustGetEnv(key string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		panic(fmt.Sprintf("required environment variable %q is not set", key))
	}
	return v
}

func getEnvOrDefault(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func mustGetEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		panic(fmt.Sprintf("environment variable %q must be an integer: %v", key, err))
	}
	return n
}
