package services

import (
	"context"
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

	// In a real implementation, you would call CoinGecko here.
	// For this prototype, we'll return hardcoded live-ish prices.
	// E.g., GET https://api.coingecko.com/api/v3/simple/price?ids=binancecoin,tether,tron&vs_currencies=usd
	
	newPrices := map[string]float64{
		"BNB":  590.50,
		"USDT": 1.00,
		"TRX":  0.135,
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
