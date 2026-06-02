package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
)

type Encryptor struct{ aead cipher.AEAD }

func New(b64Key string) (*Encryptor, error) {
	key, err := base64.StdEncoding.DecodeString(b64Key)
	if err != nil || len(key) != 32 {
		return nil, errors.New("INCOME_ENCRYPTION_KEY must be base64-encoded 32 bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Encryptor{aead: aead}, nil
}

func (e *Encryptor) EncryptIncome(amount int64) (string, error) {
	nonce := make([]byte, e.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	plain := []byte(strconv.FormatInt(amount, 10))
	ct := e.aead.Seal(nil, nonce, plain, nil)
	return base64.StdEncoding.EncodeToString(append(nonce, ct...)), nil
}

func (e *Encryptor) DecryptIncome(b64 string) (int64, error) {
	raw, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return 0, err
	}
	ns := e.aead.NonceSize()
	if len(raw) < ns {
		return 0, errors.New("ciphertext too short")
	}
	plain, err := e.aead.Open(nil, raw[:ns], raw[ns:], nil)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(string(plain), 10, 64)
}

func MaskIncome(amount int64) string {
	switch {
	case amount >= 1_000_000:
		return fmt.Sprintf("Rp %djt", amount/1_000_000)
	case amount >= 1_000:
		return fmt.Sprintf("Rp %drb", amount/1_000)
	default:
		return fmt.Sprintf("Rp %d", amount)
	}
}
