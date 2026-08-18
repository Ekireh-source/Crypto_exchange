// Package bsc implements the BlockchainAdapter interface for the
// Binance Smart Chain (BSC) and any EVM-compatible network.
//
// It uses github.com/ethereum/go-ethereum under the hood — the same library
// already present in go.mod.
package bsc

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/ethclient"
)

// Client wraps an ethclient connection to a BSC (or any EVM) node.
type Client struct {
	rpc     *ethclient.Client
	chainID *big.Int
	network string // "BSC" or "BSC_TESTNET"
}

// NewClient dials the given RPC endpoint and returns a ready-to-use Client.
// rpcURL examples:
//   - Mainnet: "https://bsc-dataseed.binance.org/"
//   - Testnet: "https://data-seed-prebsc-1-s1.binance.org:8545/"
func NewClient(ctx context.Context, rpcURL string, network string) (*Client, error) {
	rpc, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return nil, fmt.Errorf("bsc: dialing %s: %w", rpcURL, err)
	}

	chainID, err := rpc.ChainID(ctx)
	if err != nil {
		rpc.Close()
		return nil, fmt.Errorf("bsc: fetching chain ID: %w", err)
	}

	return &Client{
		rpc:     rpc,
		chainID: chainID,
		network: network,
	}, nil
}

// Network implements blockchain.BlockchainAdapter.
func (c *Client) Network() string {
	return c.network
}

// Close releases the underlying RPC connection.
func (c *Client) Close() {
	c.rpc.Close()
}

// ChainID returns the network chain ID (56 for BSC mainnet, 97 for testnet).
func (c *Client) ChainID() *big.Int {
	return c.chainID
}

// EthClient exposes the underlying go-ethereum client for advanced usage.
func (c *Client) EthClient() *ethclient.Client {
	return c.rpc
}
