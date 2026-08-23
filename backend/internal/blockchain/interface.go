// Package blockchain defines the adapter interface that every supported
// blockchain must implement.  A BlockchainAdapter abstracts away the
// network-specific details of:
//   - generating wallets (address + key derivation)
//   - reading balances (native coin and ERC/BEP/TRC-20 tokens)
//   - signing and broadcasting transactions
//   - monitoring inbound deposits
//
// Adding a new chain (e.g. Solana, Bitcoin) means writing a new type that
// satisfies this interface — no changes to service layer code required.
package blockchain

import (
	"context"
	"math/big"
	"time"
)

// IncomingTx describes a deposit detected by the monitor goroutine.
type IncomingTx struct {
	TxHash      string
	FromAddress string
	ToAddress   string
	Amount      *big.Int   // in the smallest unit (wei, sun, satoshi, …)
	Decimals    int
	Asset       string     // "NATIVE" for the chain's native coin; contract address for tokens (BEP-20/TRC-20)
	Network     string     // e.g. "BSC" or "TRON"
	BlockNumber int64
	Timestamp   time.Time
}

// FeeEstimate holds the estimated transaction cost.
type FeeEstimate struct {
	// NativeAmount is the fee denominated in the chain's native gas coin.
	// e.g. 0.0003 BNB for a BSC transfer.
	NativeAmount *big.Float
	// NativeSymbol is the gas coin ticker, e.g. "BNB" or "TRX".
	NativeSymbol string
	// USDAmount is an approximate USD equivalent (may be nil if price unavailable).
	USDAmount *big.Float
}

// BlockchainAdapter is the interface every chain adapter must implement.
type BlockchainAdapter interface {
	// Network returns the canonical network name, e.g. "BSC" or "TRON".
	Network() string

	// GenerateWallet creates a fresh keypair.
	// It returns the public address and the raw hex-encoded private key.
	// The caller is responsible for encrypting the private key before storage.
	GenerateWallet() (address string, privateKeyHex string, err error)

	// GetNativeBalance returns the chain's native coin balance (BNB, TRX, …)
	// for the given address, expressed as a human-readable decimal (e.g. 1.5).
	GetNativeBalance(ctx context.Context, address string) (*big.Float, error)

	// GetTokenBalance returns the ERC/BEP/TRC-20 token balance for the given
	// address and token contract, expressed as a human-readable decimal.
	GetTokenBalance(ctx context.Context, address, contractAddress string, decimals int) (*big.Float, error)

	// SendNative signs and broadcasts a native coin transfer.
	// amount is in human-readable units (e.g. 1.5 BNB, not wei).
	// Returns the transaction hash.
	SendNative(ctx context.Context, fromPrivKeyHex, toAddress string, amount *big.Float) (txHash string, err error)

	// SendToken signs and broadcasts a token transfer to the given contract.
	// amount is in human-readable units.
	SendToken(ctx context.Context, fromPrivKeyHex, toAddress, contractAddress string, amount *big.Float, decimals int) (txHash string, err error)

	// EstimateFee returns the estimated fee for a transfer on this network.
	// assetIsToken indicates whether the transfer involves a token (higher gas)
	// or a plain native transfer.
	EstimateFee(ctx context.Context, assetIsToken bool) (*FeeEstimate, error)

	// GetIncomingTransactions returns inbound transfers received at address
	// since the given block number. Implementations should cap the range to
	// avoid oversized RPC responses.
	GetIncomingTransactions(ctx context.Context, address string, fromBlock int64) ([]IncomingTx, error)

	// GetTxStatus checks whether a previously broadcast transaction has been
	// confirmed on-chain. Returns (true, nil) if confirmed, (false, nil) if
	// still pending, and (false, err) if the lookup itself failed.
	GetTxStatus(ctx context.Context, txHash string) (confirmed bool, err error)
}

// DepositScanner is an interface for background jobs that watch the blockchain for deposits.
type DepositScanner interface {
	// AddAddress tells the scanner to start watching a newly generated deposit address.
	AddAddress(addr string)
}
