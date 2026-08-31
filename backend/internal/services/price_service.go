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
	hub         *WSHub
	lastFetched time.Time
	mu          sync.RWMutex
}

// NewPriceService creates a PriceService.
func NewPriceService(apiKey string, hub *WSHub) *PriceService {
	return &PriceService{
		apiKey: apiKey,
		cache:  make(map[string]float64),
		hub:    hub,
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

// StartPriceTicker runs in the background. It updates the real prices every 60 seconds
// from CoinGecko, but broadcasts simulated micro-fluctuations (+/- 0.05%) every 3 seconds
// to the WebSocket Hub so the UI feels hyper-active and live.
func (s *PriceService) StartPriceTicker(ctx context.Context) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	// Initial fetch
	_, _ = s.GetPrices(ctx, nil)

	// Keep track of the current simulated prices
	currentSimulated := make(map[string]float64)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Ensure we fetch real prices occasionally by calling GetPrices
			realPrices, _ := s.GetPrices(ctx, nil)

			if len(realPrices) == 0 {
				continue
			}

			// Add micro-volatility
			simulated := make(map[string]float64)
			for symbol, realPrice := range realPrices {
				// Initialize if empty
				if _, ok := currentSimulated[symbol]; !ok {
					currentSimulated[symbol] = realPrice
				}

				// Drift back towards the real price if we deviated too far (> 0.2%)
				deviation := (currentSimulated[symbol] - realPrice) / realPrice
				
				// Calculate random fluctuation (between -0.05% and +0.05%)
				// Use time.Now().UnixNano() for pseudo-randomness without math/rand overhead
				randSeed := float64(time.Now().UnixNano()%100) / 100.0 // 0.0 to 1.0
				fluctuation := (randSeed - 0.5) * 0.001 // -0.0005 to +0.0005

				if deviation > 0.002 { // Too high, force down
					fluctuation -= 0.0005
				} else if deviation < -0.002 { // Too low, force up
					fluctuation += 0.0005
				}

				// Stablecoins like USDT should have minimal to zero volatility
				if symbol == "USDT" || symbol == "USDC" {
					fluctuation = 0
					currentSimulated[symbol] = realPrice
				} else {
					currentSimulated[symbol] = currentSimulated[symbol] * (1.0 + fluctuation)
				}

				simulated[symbol] = currentSimulated[symbol]
			}

			// Broadcast to WebSockets
			if s.hub != nil {
				s.hub.Broadcast <- simulated
			}
		}
	}
}
