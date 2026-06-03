// Package encryption provides AES-256-GCM symmetric encryption for the
// income field. Output is base64-encoded `nonce || ciphertext+tag` so it
// fits the `text` column declared in migration 0002 (matches the PRD's
// `income_encrypted TEXT` shape).
//
// Why AES-256-GCM:
//   - Authenticated encryption — tampering with the DB value yields a
//     decryption error rather than silently corrupting the income.
//   - Standard library only, no third-party crypto.
//   - Key is 32 raw bytes; we accept it as a hex string from config so it
//     survives env-var transport.
package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

// Cipher is safe for concurrent use after construction — aes.Block and
// cipher.AEAD are read-only.
type Cipher struct {
	aead cipher.AEAD
}

// NewCipherFromHex builds a Cipher from a 64-character hex string (32 raw
// bytes). The config layer is the only caller; integration tests should
// use the same key the API is configured with.
func NewCipherFromHex(hexKey string) (*Cipher, error) {
	raw, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("decoding hex key: %w", err)
	}
	if len(raw) != 32 {
		return nil, fmt.Errorf("encryption key must be 32 bytes, got %d", len(raw))
	}
	block, err := aes.NewCipher(raw)
	if err != nil {
		return nil, fmt.Errorf("creating aes block: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("creating gcm: %w", err)
	}
	return &Cipher{aead: aead}, nil
}

// Encrypt returns base64(nonce || ciphertext+tag). The nonce is fresh per
// call (crypto/rand) so two encryptions of the same plaintext produce
// different outputs — table-driven tests must encrypt fresh, not compare
// against a hard-coded golden value.
func (c *Cipher) Encrypt(plaintext []byte) (string, error) {
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("reading nonce: %w", err)
	}
	sealed := c.aead.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// Decrypt reverses Encrypt. Returns a typed error if the ciphertext is
// shorter than the nonce or fails authentication — callers should treat
// both as "corrupted / wrong key" and never expose the raw error to users.
var ErrCipherTooShort = errors.New("ciphertext shorter than nonce")

func (c *Cipher) Decrypt(b64 string) ([]byte, error) {
	raw, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, fmt.Errorf("decoding base64: %w", err)
	}
	ns := c.aead.NonceSize()
	if len(raw) < ns {
		return nil, ErrCipherTooShort
	}
	nonce, sealed := raw[:ns], raw[ns:]
	plain, err := c.aead.Open(nil, nonce, sealed, nil)
	if err != nil {
		return nil, fmt.Errorf("aead open: %w", err)
	}
	return plain, nil
}
