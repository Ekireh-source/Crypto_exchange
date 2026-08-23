package tron

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"

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
	Visible          bool   `json:"visible"`
}

// triggerSmartContractResp holds the unsigned transaction from TronGrid.
type triggerSmartContractResp struct {
	Result struct {
		Result  bool   `json:"result"`
		Code    string `json:"code"`
		Message string `json:"message"`
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
		Visible      bool   `json:"visible"`
	}
	var txResp map[string]interface{}
	if err := c.post(ctx, "/wallet/createtransaction", createTxReq{
		OwnerAddress: fromAddr,
		ToAddress:    toAddress,
		Amount:       sun.Int64(),
		Visible:      true,
	}, &txResp); err != nil {
		return "", fmt.Errorf("tron: creating TRX tx: %w", err)
	}

	if errMsg, ok := txResp["Error"].(string); ok && errMsg != "" {
		return "", fmt.Errorf("tron: createtransaction failed: %s", errMsg)
	}

	txID, err := c.signAndBroadcast(ctx, txResp, fromPrivKeyHex)
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
	// Parameter format: 32-byte address (left zero-padded) + 32-byte uint256 (left zero-padded).
	addrHex, err := tronAddressToHex(toAddress)
	if err != nil {
		return "", fmt.Errorf("tron: encoding recipient address: %w", err)
	}
	// addrHex is a 40-char hex string (20 bytes). Treat it as a big integer and
	// print as 64-char hex to get the ABI-required left-zero-padded 32-byte word.
	addrInt := new(big.Int)
	addrInt.SetString(addrHex, 16)
	param := fmt.Sprintf("%064x%064x", addrInt, baseAmount)

	var txResp triggerSmartContractResp
	if err := c.post(ctx, "/wallet/triggersmartcontract", triggerSmartContractReq{
		OwnerAddress:     fromAddr,
		ContractAddress:  contractAddress,
		FunctionSelector: "transfer(address,uint256)",
		Parameter:        param,
		FeeLimit:         40_000_000, // 40 TRX max fee (in SUN)
		Visible:          true,
	}, &txResp); err != nil {
		return "", fmt.Errorf("tron: triggering TRC-20 transfer: %w", err)
	}

	if txResp.Result.Code != "" {
		// TronGrid uses hex encoding for the error message in triggersmartcontract
		msgBytes, _ := hex.DecodeString(txResp.Result.Message)
		return "", fmt.Errorf("tron: triggersmartcontract failed: %s - %s", txResp.Result.Code, string(msgBytes))
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

// signAndBroadcast signs a raw TRON transaction and broadcasts it.
//
// TRON signing protocol:
//  1. TronGrid returns raw_data_hex: the hex-encoded raw Protobuf bytes of the tx.
//  2. SHA-256 hash the raw bytes (NOT double-SHA256, NOT Ethereum-prefixed).
//  3. Sign the 32-byte hash with secp256k1 ECDSA — same key algorithm as Ethereum.
//  4. The signature is 65 bytes [r(32)|s(32)|v(1)] where v is 0 or 1.
//  5. Broadcast the original transaction JSON with the "signature" field appended.
func (c *Client) signAndBroadcast(ctx context.Context, tx map[string]interface{}, privKeyHex string) (string, error) {
	// 1. Extract raw_data_hex from the transaction map.
	rawDataHexIface, ok := tx["raw_data_hex"]
	if !ok {
		return "", fmt.Errorf("tron: transaction missing raw_data_hex field")
	}
	rawDataHex, ok := rawDataHexIface.(string)
	if !ok {
		return "", fmt.Errorf("tron: raw_data_hex is not a string")
	}

	// 2. Hex-decode raw bytes.
	rawBytes, err := hex.DecodeString(rawDataHex)
	if err != nil {
		return "", fmt.Errorf("tron: decoding raw_data_hex: %w", err)
	}

	// 3. SHA-256 of the raw bytes → 32-byte hash (TRON protocol).
	hashArr := sha256.Sum256(rawBytes)
	hash := hashArr[:]

	// 4. Parse private key and sign.
	privKeyHex = strings.TrimPrefix(privKeyHex, "0x")
	privKeyBytes, err := hex.DecodeString(privKeyHex)
	if err != nil {
		return "", fmt.Errorf("tron: decoding private key: %w", err)
	}
	privKey, err := crypto.ToECDSA(privKeyBytes)
	if err != nil {
		return "", fmt.Errorf("tron: parsing private key: %w", err)
	}

	// crypto.Sign produces [r(32)|s(32)|v(1)]. TRON expects exactly this format.
	sig, err := crypto.Sign(hash, privKey)
	if err != nil {
		return "", fmt.Errorf("tron: signing tx: %w", err)
	}

	// 5. Append the hex-encoded signature to the transaction and broadcast.
	tx["signature"] = []string{hex.EncodeToString(sig)}

	var bcastResp broadcastResp
	if err := c.post(ctx, "/wallet/broadcasttransaction", tx, &bcastResp); err != nil {
		return "", fmt.Errorf("tron: broadcasting tx: %w", err)
	}
	if !bcastResp.Result {
		return "", fmt.Errorf("tron: broadcast failed — code: %s, msg: %s", bcastResp.Code, bcastResp.Msg)
	}

	return bcastResp.TxID, nil
}

// tronAddressToHex converts a Base58Check TRON address to its 20-byte hex
// representation (without the 0x41 prefix) for ABI encoding.
//
// TRON Base58Check → bytes:
//   - Decode Base58 → 25 bytes: [0x41 | 20-byte address | 4-byte checksum]
//   - Strip the 0x41 network prefix → 20 bytes of address
//   - Return as 40-char lowercase hex (no 0x prefix, for ABI left-pad later)
func tronAddressToHex(addr string) (string, error) {
	decoded, err := base58Decode(addr)
	if err != nil {
		return "", fmt.Errorf("tron: base58 decode of address %q: %w", addr, err)
	}
	if len(decoded) != 25 {
		return "", fmt.Errorf("tron: unexpected decoded address length %d (want 25)", len(decoded))
	}
	// decoded[0]   = 0x41  (mainnet prefix)
	// decoded[1:21] = 20-byte raw address
	// decoded[21:25] = 4-byte checksum
	addrBytes := decoded[1:21]
	return fmt.Sprintf("%x", addrBytes), nil
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

// getTxByIDResp is a partial shape of the TronGrid gettransactionbyid response.
type getTxByIDResp struct {
	TxID string `json:"txID"` // empty if the transaction is not found
}

// GetTxStatus checks whether a TRON transaction is confirmed on-chain.
// It calls GET /wallet/gettransactionbyid. A non-empty txID in the response
// means the node has a record of the transaction (i.e., it is confirmed).
func (c *Client) GetTxStatus(ctx context.Context, txHash string) (bool, error) {
	path := "/wallet/gettransactionbyid?value=" + txHash
	var resp getTxByIDResp
	if err := c.get(ctx, path, &resp); err != nil {
		return false, fmt.Errorf("tron: GetTxStatus %s: %w", txHash, err)
	}
	return resp.TxID != "", nil
}

