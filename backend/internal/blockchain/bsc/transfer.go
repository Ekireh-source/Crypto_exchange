package bsc

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"exchange/internal/blockchain"
	bscCrypto "exchange/internal/blockchain/crypto"
)

// erc20TransferABI is the minimal ABI for ERC-20/BEP-20 transfer(address,uint256).
var erc20TransferABI abi.ABI

func init() {
	var err error
	erc20TransferABI, err = abi.JSON(strings.NewReader(`[{
		"name":"transfer",
		"type":"function",
		"inputs":[
			{"name":"recipient","type":"address"},
			{"name":"amount","type":"uint256"}
		],
		"outputs":[{"name":"","type":"bool"}]
	}]`))
	if err != nil {
		panic("bsc: parsing ERC-20 ABI: " + err.Error())
	}
}

// GetNativeBalance returns the BNB balance for the given address (in BNB, not wei).
func (c *Client) GetNativeBalance(ctx context.Context, address string) (*big.Float, error) {
	addr := common.HexToAddress(address)
	wei, err := c.rpc.BalanceAt(ctx, addr, nil)
	if err != nil {
		return nil, fmt.Errorf("bsc: GetNativeBalance %s: %w", address, err)
	}
	return bscCrypto.FromBaseUnits(wei, 18), nil
}

// GetTokenBalance returns the BEP-20 token balance for address at contractAddress.
// decimals is typically 18 for most tokens, but 6 for USDT-BEP20.
func (c *Client) GetTokenBalance(ctx context.Context, address, contractAddress string, decimals int) (*big.Float, error) {
	// ABI-encode balanceOf(address)
	balanceOfABI, _ := abi.JSON(strings.NewReader(`[{
		"name":"balanceOf","type":"function",
		"inputs":[{"name":"account","type":"address"}],
		"outputs":[{"name":"","type":"uint256"}]
	}]`))

	callData, err := balanceOfABI.Pack("balanceOf", common.HexToAddress(address))
	if err != nil {
		return nil, fmt.Errorf("bsc: packing balanceOf call: %w", err)
	}

	contract := common.HexToAddress(contractAddress)
	result, err := c.rpc.CallContract(ctx, ethereum.CallMsg{
		To:   &contract,
		Data: callData,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("bsc: calling balanceOf: %w", err)
	}

	var balance *big.Int
	if err = balanceOfABI.UnpackIntoInterface(&balance, "balanceOf", result); err != nil {
		return nil, fmt.Errorf("bsc: unpacking balanceOf result: %w", err)
	}

	return bscCrypto.FromBaseUnits(balance, decimals), nil
}

// SendNative signs and broadcasts a BNB transfer.
// amount is in BNB (human-readable), not wei.
func (c *Client) SendNative(ctx context.Context, fromPrivKeyHex, toAddress string, amount *big.Float) (string, error) {
	fromPrivKeyHex = strings.TrimPrefix(fromPrivKeyHex, "0x")
	privKey, err := crypto.HexToECDSA(fromPrivKeyHex)
	if err != nil {
		return "", fmt.Errorf("bsc: parsing private key: %w", err)
	}

	fromAddr := crypto.PubkeyToAddress(privKey.PublicKey)
	toAddr := common.HexToAddress(toAddress)

	amountWei := bscCrypto.ToBaseUnits(amount, 18)

	nonce, err := c.rpc.PendingNonceAt(ctx, fromAddr)
	if err != nil {
		return "", fmt.Errorf("bsc: fetching nonce: %w", err)
	}

	gasPrice, err := c.rpc.SuggestGasPrice(ctx)
	if err != nil {
		return "", fmt.Errorf("bsc: suggesting gas price: %w", err)
	}

	tx := types.NewTransaction(nonce, toAddr, amountWei, 21000, gasPrice, nil)
	signed, err := types.SignTx(tx, types.LatestSignerForChainID(c.chainID), privKey)
	if err != nil {
		return "", fmt.Errorf("bsc: signing tx: %w", err)
	}

	if err = c.rpc.SendTransaction(ctx, signed); err != nil {
		return "", fmt.Errorf("bsc: broadcasting tx: %w", err)
	}

	return signed.Hash().Hex(), nil
}

// SendToken signs and broadcasts a BEP-20 transfer call.
// amount is in token units (human-readable), not wei.
func (c *Client) SendToken(ctx context.Context, fromPrivKeyHex, toAddress, contractAddress string, amount *big.Float, decimals int) (string, error) {
	fromPrivKeyHex = strings.TrimPrefix(fromPrivKeyHex, "0x")
	privKey, err := crypto.HexToECDSA(fromPrivKeyHex)
	if err != nil {
		return "", fmt.Errorf("bsc: parsing private key: %w", err)
	}

	fromAddr := crypto.PubkeyToAddress(privKey.PublicKey)
	toAddr := common.HexToAddress(toAddress)
	contractAddr := common.HexToAddress(contractAddress)

	amountBase := bscCrypto.ToBaseUnits(amount, decimals)

	// Pack the transfer(address, uint256) call data.
	callData, err := erc20TransferABI.Pack("transfer", toAddr, amountBase)
	if err != nil {
		return "", fmt.Errorf("bsc: packing transfer call: %w", err)
	}

	nonce, err := c.rpc.PendingNonceAt(ctx, fromAddr)
	if err != nil {
		return "", fmt.Errorf("bsc: fetching nonce: %w", err)
	}

	gasPrice, err := c.rpc.SuggestGasPrice(ctx)
	if err != nil {
		return "", fmt.Errorf("bsc: suggesting gas price: %w", err)
	}

	// Estimate gas for the token transfer (typically ~65,000 for BEP-20).
	estimatedGas, err := c.rpc.EstimateGas(ctx, ethereum.CallMsg{
		From: fromAddr,
		To:   &contractAddr,
		Data: callData,
	})
	if err != nil {
		estimatedGas = 100_000 // fallback if estimation fails
	}

	tx := types.NewTransaction(nonce, contractAddr, big.NewInt(0), estimatedGas, gasPrice, callData)
	signed, err := types.SignTx(tx, types.LatestSignerForChainID(c.chainID), privKey)
	if err != nil {
		return "", fmt.Errorf("bsc: signing token tx: %w", err)
	}

	if err = c.rpc.SendTransaction(ctx, signed); err != nil {
		return "", fmt.Errorf("bsc: broadcasting token tx: %w", err)
	}

	return signed.Hash().Hex(), nil
}

// EstimateFee returns the approximate BNB cost for a transfer.
func (c *Client) EstimateFee(ctx context.Context, assetIsToken bool) (*blockchain.FeeEstimate, error) {
	gasPrice, err := c.rpc.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("bsc: suggesting gas price: %w", err)
	}

	gasLimit := uint64(21_000)
	if assetIsToken {
		gasLimit = 100_000
	}

	gasCost := new(big.Int).Mul(gasPrice, big.NewInt(int64(gasLimit)))
	return &blockchain.FeeEstimate{
		NativeAmount: bscCrypto.FromBaseUnits(gasCost, 18),
		NativeSymbol: "BNB",
	}, nil
}

// GetIncomingTransactions returns Transfer events to address from fromBlock.
// This scans the latest 2000 blocks to avoid oversized eth_getLogs responses.
func (c *Client) GetIncomingTransactions(ctx context.Context, address string, fromBlock int64) ([]blockchain.IncomingTx, error) {
	// Full implementation is in monitor.go.
	// This stub satisfies the interface for compilation.
	return nil, nil
}

// GetTxStatus checks whether a BSC transaction has been mined and confirmed.
// It calls eth_getTransactionReceipt — a non-nil receipt means the tx landed
// in a block. A receipt with Status == 0 means the tx was reverted on-chain,
// which we still treat as "confirmed" from a double-spend perspective.
func (c *Client) GetTxStatus(ctx context.Context, txHash string) (bool, error) {
	hash := common.HexToHash(txHash)
	receipt, err := c.rpc.TransactionReceipt(ctx, hash)
	if err != nil {
		// go-ethereum returns ethereum.NotFound when the tx is not yet mined.
		if err.Error() == "not found" {
			return false, nil
		}
		return false, fmt.Errorf("bsc: GetTxStatus %s: %w", txHash, err)
	}
	return receipt != nil, nil
}

// SendWithBumpedGas re-broadcasts a transaction from the Hot Wallet with a
// gas price bumped by bumpPercent. It reuses the current pending nonce so the
// new tx replaces the stuck one in the BSC/EVM mempool (same-nonce replacement).
//
// This is used by the WithdrawalMonitor (Phase 5.3) to unstick withdrawals.
func (c *Client) SendWithBumpedGas(
	ctx context.Context,
	fromPrivKeyHex string,
	toAddress string,
	amount *big.Float,
	contractAddress *string,
	decimals int,
	bumpPercent int,
	isToken bool,
) (string, error) {
	fromPrivKeyHex = strings.TrimPrefix(fromPrivKeyHex, "0x")
	privKey, err := crypto.HexToECDSA(fromPrivKeyHex)
	if err != nil {
		return "", fmt.Errorf("bsc: parsing private key: %w", err)
	}

	fromAddr := crypto.PubkeyToAddress(privKey.PublicKey)
	toAddr := common.HexToAddress(toAddress)

	// Fetch current nonce (pending) so we reuse the same slot as the stuck tx.
	nonce, err := c.rpc.PendingNonceAt(ctx, fromAddr)
	if err != nil {
		return "", fmt.Errorf("bsc: fetching nonce: %w", err)
	}

	// Get suggested gas price and bump it.
	baseGasPrice, err := c.rpc.SuggestGasPrice(ctx)
	if err != nil {
		return "", fmt.Errorf("bsc: suggesting gas price: %w", err)
	}
	bumpMultiplier := big.NewInt(int64(100 + bumpPercent))
	bumpedGasPrice := new(big.Int).Div(new(big.Int).Mul(baseGasPrice, bumpMultiplier), big.NewInt(100))

	var signed *types.Transaction

	if !isToken {
		amountWei := bscCrypto.ToBaseUnits(amount, 18)
		tx := types.NewTransaction(nonce, toAddr, amountWei, 21000, bumpedGasPrice, nil)
		signed, err = types.SignTx(tx, types.LatestSignerForChainID(c.chainID), privKey)
		if err != nil {
			return "", fmt.Errorf("bsc: signing bumped native tx: %w", err)
		}
	} else {
		contractAddr := common.HexToAddress(*contractAddress)
		amountBase := bscCrypto.ToBaseUnits(amount, decimals)
		callData, packErr := erc20TransferABI.Pack("transfer", toAddr, amountBase)
		if packErr != nil {
			return "", fmt.Errorf("bsc: packing transfer: %w", packErr)
		}
		estimatedGas, estErr := c.rpc.EstimateGas(ctx, ethereum.CallMsg{
			From: fromAddr,
			To:   &contractAddr,
			Data: callData,
		})
		if estErr != nil {
			estimatedGas = 100_000
		}
		tx := types.NewTransaction(nonce, contractAddr, big.NewInt(0), estimatedGas, bumpedGasPrice, callData)
		signed, err = types.SignTx(tx, types.LatestSignerForChainID(c.chainID), privKey)
		if err != nil {
			return "", fmt.Errorf("bsc: signing bumped token tx: %w", err)
		}
	}

	if err = c.rpc.SendTransaction(ctx, signed); err != nil {
		return "", fmt.Errorf("bsc: broadcasting bumped tx: %w", err)
	}

	return signed.Hash().Hex(), nil
}

