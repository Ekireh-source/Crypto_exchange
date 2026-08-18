// Package crypto provides shared cryptographic utilities used by blockchain
// adapters: AES-256-GCM encryption for private key storage, and helpers for
// converting between human-readable amounts and on-chain integer amounts.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/big"
)

// EncryptPrivateKey encrypts a hex-encoded private key using AES-256-GCM.
// keyHex must be a 64-character hex string (32 bytes).
// The returned string is hex-encoded ciphertext that is safe to store in the DB.
func EncryptPrivateKey(privKeyHex, keyHex string) (string, error) {
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return "", fmt.Errorf("decoding encryption key: %w", err)
	}
	if len(key) != 32 {
		return "", errors.New("encryption key must be exactly 32 bytes (64 hex chars)")
	}

	plaintext := []byte(privKeyHex)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("creating AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("creating GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generating nonce: %w", err)
	}

	// Prepend nonce to ciphertext so we can extract it during decryption.
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return hex.EncodeToString(ciphertext), nil
}

// DecryptPrivateKey reverses EncryptPrivateKey.
// Returns the original hex-encoded private key.
func DecryptPrivateKey(encryptedHex, keyHex string) (string, error) {
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return "", fmt.Errorf("decoding encryption key: %w", err)
	}

	data, err := hex.DecodeString(encryptedHex)
	if err != nil {
		return "", fmt.Errorf("decoding ciphertext: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("creating AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("creating GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypting: %w", err)
	}

	return string(plaintext), nil
}

// ToBaseUnits converts a human-readable amount (e.g. "1.5") to the integer
// base-unit representation used on-chain (e.g. 1500000000000000000 for 18
// decimals).
func ToBaseUnits(amount *big.Float, decimals int) *big.Int {
	multiplier := new(big.Float).SetInt(Pow10(decimals))
	result := new(big.Float).Mul(amount, multiplier)

	intVal, _ := result.Int(nil)
	return intVal
}

// FromBaseUnits converts an on-chain integer amount to a human-readable
// big.Float (e.g. 1500000000000000000 wei → 1.5 ETH/BNB).
func FromBaseUnits(amount *big.Int, decimals int) *big.Float {
	divisor := new(big.Float).SetInt(Pow10(decimals))
	return new(big.Float).Quo(new(big.Float).SetInt(amount), divisor)
}

// Pow10 returns 10^n as a *big.Int.
func Pow10(n int) *big.Int {
	result := big.NewInt(1)
	ten := big.NewInt(10)
	for i := 0; i < n; i++ {
		result.Mul(result, ten)
	}
	return result
}
