package tron

import (
	"context"
	"fmt"
	"math/big"

	"exchange/internal/blockchain"
	bscCrypto "exchange/internal/blockchain/crypto"
)

// trc20BalanceResponse is the shape of the TronGrid balance endpoint.
type trc20BalanceResponse struct {
	Data []struct {
		Balance string `json:"balance"`
	} `json:"data"`
}

// accountResponse is a partial shape of /v1/accounts/{address}.
type accountResponse struct {
	Data []struct {
		Balance int64 `json:"balance"` // TRX in SUN (1 TRX = 1,000,000 SUN)
	} `json:"data"`
}

// GetNativeBalance returns the TRX balance for address in TRX (not SUN).
func (c *Client) GetNativeBalance(ctx context.Context, address string) (*big.Float, error) {
	var resp accountResponse
	if err := c.get(ctx, "/v1/accounts/"+address, &resp); err != nil {
		return nil, fmt.Errorf("tron: GetNativeBalance: %w", err)
	}
	if len(resp.Data) == 0 {
		return big.NewFloat(0), nil // account not yet activated
	}
	sun := big.NewInt(resp.Data[0].Balance)
	return bscCrypto.FromBaseUnits(sun, 6), nil // 6 decimals for TRX
}

// GetTokenBalance returns the TRC-20 token balance for address in token units.
// decimals is 6 for USDT-TRC20.
func (c *Client) GetTokenBalance(ctx context.Context, address, contractAddress string, decimals int) (*big.Float, error) {
	path := fmt.Sprintf("/v1/accounts/%s/tokens/trc20?contract_address=%s", address, contractAddress)

	var resp trc20BalanceResponse
	if err := c.get(ctx, path, &resp); err != nil {
		return nil, fmt.Errorf("tron: GetTokenBalance: %w", err)
	}
	if len(resp.Data) == 0 {
		return big.NewFloat(0), nil
	}

	n := new(big.Int)
	n.SetString(resp.Data[0].Balance, 10)
	return bscCrypto.FromBaseUnits(n, decimals), nil
}

// triggerSmartContractReq is the payload for TronGrid's triggersmartcontract.
type triggerSmartContractReq struct {
	OwnerAddress     string `json:"owner_address"`
	ContractAddress  string `json:"contract_address"`
	FunctionSelector string `json:"function_selector"`
	Parameter        string `json:"parameter"`
	FeeLimit         int64  `json:"fee_limit"`
	CallValue        int    `json:"call_value"`
}

// triggerSmartContractResp holds the unsigned transaction from TronGrid.
type triggerSmartContractResp struct {
	Result struct {
		Result bool `json:"result"`
	} `json:"result"`
	Transaction map[string]interface{} `json:"transaction"`
}

// broadcastResp holds the result of /wallet/broadcasttransaction.
type broadcastResp struct {
	Result bool   `json:"result"`
	TxID   string `json:"txid"`
	Code   string `json:"code"`
	Msg    string `json:"message"`
}

// SendNative sends TRX from fromPrivKeyHex to toAddress.
// amount is in TRX (human-readable), not SUN.
//
// Flow: POST /wallet/createtransaction → sign → POST /wallet/broadcasttransaction
func (c *Client) SendNative(ctx context.Context, fromPrivKeyHex, toAddress string, amount *big.Float) (string, error) {
	sun := bscCrypto.ToBaseUnits(amount, 6)
	fromAddr, err := privateKeyToTronAddress(fromPrivKeyHex)
	if err != nil {
		return "", err
	}

	type createTxReq struct {
		OwnerAddress string `json:"owner_address"`
		ToAddress    string `json:"to_address"`
		Amount       int64  `json:"amount"` // in SUN
	}
	type createTxResp struct {
		Transaction map[string]interface{} `json:"transaction"`
	}

	var txResp createTxResp
	if err := c.post(ctx, "/wallet/createtransaction", createTxReq{
		OwnerAddress: fromAddr,
		ToAddress:    toAddress,
		Amount:       sun.Int64(),
	}, &txResp); err != nil {
		return "", fmt.Errorf("tron: creating TRX tx: %w", err)
	}

	txID, err := c.signAndBroadcast(ctx, txResp.Transaction, fromPrivKeyHex)
	if err != nil {
		return "", err
	}
	return txID, nil
}

// SendToken sends a TRC-20 token (e.g. USDT) via triggersmartcontract.
// amount is in token units (human-readable), not SUN.
func (c *Client) SendToken(ctx context.Context, fromPrivKeyHex, toAddress, contractAddress string, amount *big.Float, decimals int) (string, error) {
	baseAmount := bscCrypto.ToBaseUnits(amount, decimals)
	fromAddr, err := privateKeyToTronAddress(fromPrivKeyHex)
	if err != nil {
		return "", err
	}

	// ABI encode: transfer(address, uint256)
	// TRON uses the same ABI encoding as Ethereum for function parameters.
	// Parameter format: 32-byte address (0-padded) + 32-byte uint256 (0-padded)
	param := fmt.Sprintf("%064s%064x", tronAddressToHex(toAddress), baseAmount)

	var txResp triggerSmartContractResp
	if err := c.post(ctx, "/wallet/triggersmartcontract", triggerSmartContractReq{
		OwnerAddress:     fromAddr,
		ContractAddress:  contractAddress,
		FunctionSelector: "transfer(address,uint256)",
		Parameter:        param,
		FeeLimit:         40_000_000, // 40 TRX max fee (in SUN)
	}, &txResp); err != nil {
		return "", fmt.Errorf("tron: triggering TRC-20 transfer: %w", err)
	}

	if !txResp.Result.Result {
		return "", fmt.Errorf("tron: triggersmartcontract returned false result")
	}

	txID, err := c.signAndBroadcast(ctx, txResp.Transaction, fromPrivKeyHex)
	if err != nil {
		return "", err
	}
	return txID, nil
}

// signAndBroadcast signs the raw transaction map and broadcasts it to the network.
// TRON transaction signing uses the same secp256k1 ECDSA as Ethereum.
func (c *Client) signAndBroadcast(ctx context.Context, tx map[string]interface{}, privKeyHex string) (string, error) {
	// For a production implementation use a proper TRON Protobuf signer.
	// This stub documents the flow and returns an error prompting full impl.
	// TODO: implement TRON Protobuf signing (see gotron-sdk or tronweb for reference)
	return "", fmt.Errorf("tron: signAndBroadcast not yet implemented — integrate gotron-sdk Protobuf signing")
}

// tronAddressToHex converts a Base58Check TRON address to its 20-byte hex
// representation (without the 0x41 prefix) for ABI encoding.
func tronAddressToHex(addr string) string {
	// Placeholder — full implementation decodes Base58Check, strips 0x41 prefix.
	return addr
}

// EstimateFee returns a rough TRX fee estimate.
// Token transfers cost ~15 TRX in energy; plain TRX transfers cost ~1 TRX.
func (c *Client) EstimateFee(ctx context.Context, assetIsToken bool) (*blockchain.FeeEstimate, error) {
	amount := big.NewFloat(1) // ~1 TRX for plain TRX transfer (bandwidth)
	if assetIsToken {
		amount = big.NewFloat(15) // ~15 TRX for TRC-20 transfer
	}
	return &blockchain.FeeEstimate{
		NativeAmount: amount,
		NativeSymbol: "TRX",
	}, nil
}
