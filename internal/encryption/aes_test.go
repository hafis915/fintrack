package encryption_test

import (
	"encoding/base64"
	"testing"

	"github.com/hafis915/fintrack/internal/encryption"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	key := base64.StdEncoding.EncodeToString(make([]byte, 32))
	enc, err := encryption.New(key)
	require.NoError(t, err)

	cipher, err := enc.EncryptIncome(8_000_000)
	require.NoError(t, err)
	require.NotEmpty(t, cipher)

	got, err := enc.DecryptIncome(cipher)
	require.NoError(t, err)
	require.Equal(t, int64(8_000_000), got)
}

func TestMaskIncome(t *testing.T) {
	require.Equal(t, "Rp 8jt", encryption.MaskIncome(8_000_000))
	require.Equal(t, "Rp 12jt", encryption.MaskIncome(12_500_000))
	require.Equal(t, "Rp 950rb", encryption.MaskIncome(950_000))
}

func TestEncryptIsNonDeterministic(t *testing.T) {
	key := base64.StdEncoding.EncodeToString(make([]byte, 32))
	enc, _ := encryption.New(key)
	a, _ := enc.EncryptIncome(1_000_000)
	b, _ := enc.EncryptIncome(1_000_000)
	require.NotEqual(t, a, b, "GCM nonce must randomize ciphertext")
}
