package bsc

import (
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
)

// GenerateWallet creates a new random secp256k1 keypair for BSC / EVM.
// Returns:
//   - address: the 0x-prefixed Ethereum-compatible address
//   - privateKeyHex: the raw 64-character hex private key (no 0x prefix)
//
// The caller MUST encrypt privateKeyHex (using blockchain/crypto.EncryptPrivateKey)
// before persisting it to the database.
func (c *Client) GenerateWallet() (address string, privateKeyHex string, err error) {
	privKey, err := crypto.GenerateKey()
	if err != nil {
		return "", "", fmt.Errorf("bsc: generating key: %w", err)
	}

	addr := crypto.PubkeyToAddress(privKey.PublicKey)
	hexKey := fmt.Sprintf("%x", crypto.FromECDSA(privKey))

	return addr.Hex(), hexKey, nil
}
