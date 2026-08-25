// Package tron implements the BlockchainAdapter interface for the TRON network.
//
// Unlike BSC/EVM networks, TRON has NO official Go SDK.  This package speaks
// to TronGrid's REST API (https://developers.tron.network/reference) using
// standard net/http.
//
// Key differences vs BSC that every developer must understand:
//   - Addresses use Base58Check encoding starting with 'T' (mainnet)
//     instead of 0x hex.  But the underlying key is still secp256k1.
//   - Fees are paid in TRX as "Energy" (for smart contracts) and "Bandwidth"
//     (for plain transfers), NOT as simple gas × gasPrice.
//   - Transactions are Protobuf-encoded, not RLP.
//   - Token transfers go through triggersmartcontract, not a direct EVM call.
package tron

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ErrNotFound is returned when the TronGrid API returns a 404 Not Found,
// indicating the account is not activated.
var ErrNotFound = errors.New("tron: account not found")

// Client talks to TronGrid's REST API.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	network    string
}

// NewClient creates a TronGrid API client.
// baseURL:  "https://api.trongrid.io"  (mainnet)
//           "https://api.shasta.trongrid.io" (testnet)
func NewClient(baseURL, apiKey, network string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		network: network,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// Network implements blockchain.BlockchainAdapter.
func (c *Client) Network() string {
	return c.network
}

// ─── internal helpers ─────────────────────────────────────────────────────────

// get performs a GET request against the TronGrid API.
func (c *Client) get(ctx context.Context, path string, out interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	if c.apiKey != "" {
		req.Header.Set("TRON-PRO-API-KEY", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("tron GET %s: %w", path, err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("tron GET %s: status %d: %s", path, resp.StatusCode, body)
	}
	return json.Unmarshal(body, out)
}

// post performs a POST request with a JSON body against the TronGrid API.
func (c *Client) post(ctx context.Context, path string, payload, out interface{}) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("TRON-PRO-API-KEY", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("tron POST %s: %w", path, err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("tron POST %s: status %d: %s", path, resp.StatusCode, body)
	}
	return json.Unmarshal(body, out)
}
