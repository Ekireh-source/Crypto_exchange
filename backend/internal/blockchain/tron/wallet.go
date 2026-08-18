package tron

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/ripemd160" //nolint:staticcheck // TRON protocol requirement
)

// GenerateWallet creates a new TRON wallet.
//
// Key derivation (same cryptography as Ethereum, different encoding):
//  1. Generate a random secp256k1 keypair.
//  2. Take the 64-byte uncompressed public key (without 0x04 prefix).
//  3. Keccak256 hash the public key → take last 20 bytes (same as Ethereum).
//  4. Prepend 0x41 (TRON mainnet prefix) → 21 bytes.
//  5. Double-SHA256 the 21 bytes → take first 4 bytes as checksum.
//  6. Append checksum to the 21-byte payload → 25 bytes.
//  7. Base58 encode → "T..." address (34 characters).
//
// Returns:
//   - address: "T..." Base58 TRON address (mainnet)
//   - privateKeyHex: 64-char hex private key (no prefix)
//
// The caller MUST encrypt privateKeyHex before persisting to the DB.
func (c *Client) GenerateWallet() (address string, privateKeyHex string, err error) {
	privKey, err := crypto.GenerateKey()
	if err != nil {
		return "", "", fmt.Errorf("tron: generating key: %w", err)
	}

	hexKey := fmt.Sprintf("%x", crypto.FromECDSA(privKey))
	addr, err := privateKeyToTronAddress(hexKey)
	if err != nil {
		return "", "", err
	}

	return addr, hexKey, nil
}

// privateKeyToTronAddress derives the TRON address from a hex private key.
func privateKeyToTronAddress(privKeyHex string) (string, error) {
	privKeyBytes, err := hex.DecodeString(privKeyHex)
	if err != nil {
		return "", fmt.Errorf("tron: decoding private key: %w", err)
	}

	// Recover the secp256k1 public key.
	privKey := secp256k1.PrivKeyFromBytes(privKeyBytes)
	pubKey := privKey.PubKey()

	// Uncompressed public key (65 bytes, starts with 0x04).
	pubBytes := pubKey.SerializeUncompressed()
	// Drop the 0x04 prefix → 64 bytes.
	pubBytes = pubBytes[1:]

	// Keccak256 → take last 20 bytes (same as Ethereum).
	hash := crypto.Keccak256(pubBytes)
	addrBytes := hash[len(hash)-20:]

	// Prepend TRON mainnet prefix 0x41.
	payload := append([]byte{0x41}, addrBytes...)

	// Base58Check encode.
	return base58CheckEncode(payload), nil
}

// base58CheckEncode implements Bitcoin-style Base58Check encoding used by TRON.
func base58CheckEncode(payload []byte) string {
	// Double SHA256 for checksum.
	h1 := sha256.Sum256(payload)
	h2 := sha256.Sum256(h1[:])
	checksum := h2[:4]

	full := append(payload, checksum...)
	return base58Encode(full)
}

const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

func base58Encode(input []byte) string {
	// Count leading zero bytes.
	leadingZeros := 0
	for _, b := range input {
		if b != 0 {
			break
		}
		leadingZeros++
	}

	// Convert big-endian byte slice to base58 digits.
	var result []byte
	n := new([32]byte)
	copy(n[32-len(input):], input)

	// Simple big-number division.
	buf := make([]byte, len(input))
	copy(buf, input)

	for len(buf) > 0 {
		var remainder int
		var newBuf []byte
		for _, b := range buf {
			digit := remainder*256 + int(b)
			if len(newBuf) > 0 || digit/58 > 0 {
				newBuf = append(newBuf, byte(digit/58))
			}
			remainder = digit % 58
		}
		result = append(result, base58Alphabet[remainder])
		buf = newBuf
	}

	// Add '1' for each leading zero byte.
	for i := 0; i < leadingZeros; i++ {
		result = append(result, base58Alphabet[0])
	}

	// Reverse.
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	_ = ripemd160.New() // keep import used
	return string(result)
}
