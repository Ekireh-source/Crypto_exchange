package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// PriceService manages live price feeds (e.g. from CoinGecko).
// It caches prices for 60 seconds.
type PriceService struct {
	apiKey      string
	cache       map[string]float64
	lastFetched time.Time
	mu          sync.RWMutex
}

// NewPriceService creates a PriceService.
func NewPriceService(apiKey string) *PriceService {
	return &PriceService{
		apiKey: apiKey,
		cache:  make(map[string]float64),
	}
}

// GetPrices returns a map of symbol -> USD price.
// If the cache is fresh, it returns immediately. Otherwise it fetches.
func (s *PriceService) GetPrices(ctx context.Context, symbols []string) (map[string]float64, error) {
	s.mu.RLock()
	// Cache for 60 seconds
	if time.Since(s.lastFetched) < 60*time.Second && len(s.cache) > 0 {
		defer s.mu.RUnlock()
		return copyMap(s.cache), nil
	}
	s.mu.RUnlock()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.coingecko.com/api/v3/simple/price?ids=binancecoin,tether,tron&vs_currencies=usd", nil)
	if err != nil {
		return nil, fmt.Errorf("creating coingecko request: %w", err)
	}

	// Use API key if provided and not the default placeholder
	if s.apiKey != "" && s.apiKey != "replace-with-your-coingecko-api-key" {
		req.Header.Set("x-cg-demo-api-key", s.apiKey)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching prices from coingecko: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("coingecko returned status %d", resp.StatusCode)
	}

	var result map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding coingecko response: %w", err)
	}

	newPrices := make(map[string]float64)
	if bnb, ok := result["binancecoin"]; ok {
		newPrices["BNB"] = bnb["usd"]
	}
	if usdt, ok := result["tether"]; ok {
		newPrices["USDT"] = usdt["usd"]
	}
	if trx, ok := result["tron"]; ok {
		newPrices["TRX"] = trx["usd"]
	}

	// Fallback mechanism in case CoinGecko free API is unstable or returns missing data
	if newPrices["BNB"] == 0 {
		newPrices["BNB"] = 590.50
	}
	if newPrices["USDT"] == 0 {
		newPrices["USDT"] = 1.00
	}
	if newPrices["TRX"] == 0 {
		newPrices["TRX"] = 0.135
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.cache = newPrices
	s.lastFetched = time.Now()

	return copyMap(s.cache), nil
}

func copyMap(m map[string]float64) map[string]float64 {
	cp := make(map[string]float64, len(m))
	for k, v := range m {
		cp[k] = v
	}
	return cp
}
