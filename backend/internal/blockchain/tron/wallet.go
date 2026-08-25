package tron

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

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
	privKeyHex = strings.TrimPrefix(privKeyHex, "0x")
	privKeyBytes, err := hex.DecodeString(privKeyHex)
	if err != nil {
		return "", fmt.Errorf("tron: decoding private key: %w", err)
	}

	privKey, err := crypto.ToECDSA(privKeyBytes)
	if err != nil {
		return "", fmt.Errorf("tron: parsing private key: %w", err)
	}

	// Uncompressed public key (65 bytes, starts with 0x04).
	pubBytes := crypto.FromECDSAPub(&privKey.PublicKey)
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
	return string(result)
}

// base58Decode reverses base58Encode. Returns the raw bytes or an error.
func base58Decode(s string) ([]byte, error) {
	n := big.NewInt(0)
	base := big.NewInt(58)

	for _, c := range s {
		idx := -1
		for i, ac := range base58Alphabet {
			if ac == c {
				idx = i
				break
			}
		}
		if idx < 0 {
			return nil, fmt.Errorf("invalid base58 character: %q", c)
		}
		n.Mul(n, base)
		n.Add(n, big.NewInt(int64(idx)))
	}

	result := n.Bytes()

	// Restore leading zeros: each '1' in input = one 0x00 byte.
	nLeading := 0
	for _, c := range s {
		if c != rune(base58Alphabet[0]) {
			break
		}
		nLeading++
	}
	return append(make([]byte, nLeading), result...), nil
}

// hexToTronAddress converts a raw hex address (as returned by TronGrid in
// native transaction contract fields) to a Base58Check TRON address.
//
// TronGrid encodes native tx addresses as 42-char hex with "41" prefix
// (e.g. "41fb1311854d10d710be5578c5c68c77c3c49103a7").
// We prepend 0x41, double-SHA256 the 21 bytes for the checksum, then Base58 encode.
func hexToTronAddress(hexAddr string) (string, error) {
	// Strip an optional leading "0x" or "41" prefix so we always work with raw bytes.
	// TronGrid returns the full 21-byte payload as 42 hex chars (no 0x prefix).
	raw, err := hex.DecodeString(hexAddr)
	if err != nil {
		return "", fmt.Errorf("tron: decoding hex address %q: %w", hexAddr, err)
	}
	// If the decoded bytes are 20 bytes (missing 0x41 prefix), prepend it.
	if len(raw) == 20 {
		raw = append([]byte{0x41}, raw...)
	}
	if len(raw) != 21 {
		return "", fmt.Errorf("tron: unexpected hex address length %d (want 21 bytes)", len(raw))
	}
	return base58CheckEncode(raw), nil
}

// keep ripemd160 import for any future use (TRON uses it internally).
var _ = ripemd160.New
