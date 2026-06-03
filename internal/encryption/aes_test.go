package encryption

import (
	"strings"
	"testing"
)

const testKey = "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff"

func TestNewCipherFromHex_RejectsBadKeys(t *testing.T) {
	cases := map[string]string{
		"not_hex":    "zz112233445566778899aabbccddeeff00112233445566778899aabbccddeeff",
		"too_short":  "00112233445566778899aabbccddeeff",
		"too_long":   testKey + "00",
		"empty":      "",
	}
	for name, key := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := NewCipherFromHex(key); err == nil {
				t.Errorf("expected error for %s key, got nil", name)
			}
		})
	}
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	c, err := NewCipherFromHex(testKey)
	if err != nil {
		t.Fatalf("NewCipherFromHex: %v", err)
	}

	in := []byte("8000000") // rupiah amount, plaintext
	sealed, err := c.Encrypt(in)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if sealed == "" {
		t.Fatal("encrypted output is empty")
	}
	if strings.Contains(sealed, "8000000") {
		t.Fatal("ciphertext contains plaintext substring — encryption broken")
	}

	out, err := c.Decrypt(sealed)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if string(out) != string(in) {
		t.Errorf("round-trip mismatch: want %q, got %q", in, out)
	}
}

func TestEncrypt_NonceUniqueAcrossCalls(t *testing.T) {
	c, err := NewCipherFromHex(testKey)
	if err != nil {
		t.Fatalf("NewCipherFromHex: %v", err)
	}

	a, _ := c.Encrypt([]byte("same"))
	b, _ := c.Encrypt([]byte("same"))
	if a == b {
		t.Fatal("two encryptions of the same plaintext produced identical output — nonce isn't randomized")
	}
}

func TestDecrypt_RejectsTampered(t *testing.T) {
	c, err := NewCipherFromHex(testKey)
	if err != nil {
		t.Fatalf("NewCipherFromHex: %v", err)
	}

	sealed, _ := c.Encrypt([]byte("9000000"))
	// Flip a character mid-ciphertext.
	tampered := []byte(sealed)
	tampered[len(tampered)/2] ^= 0x01
	if _, err := c.Decrypt(string(tampered)); err == nil {
		t.Fatal("expected decryption to fail on tampered ciphertext")
	}
}

func TestDecrypt_RejectsTooShort(t *testing.T) {
	c, err := NewCipherFromHex(testKey)
	if err != nil {
		t.Fatalf("NewCipherFromHex: %v", err)
	}
	if _, err := c.Decrypt("AAAA"); err == nil {
		t.Fatal("expected error for short ciphertext")
	}
}

func TestDecrypt_RejectsWrongKey(t *testing.T) {
	a, _ := NewCipherFromHex(testKey)
	b, _ := NewCipherFromHex("ffeeddccbbaa99887766554433221100ffeeddccbbaa99887766554433221100")

	sealed, _ := a.Encrypt([]byte("8500000"))
	if _, err := b.Decrypt(sealed); err == nil {
		t.Fatal("expected decryption with wrong key to fail")
	}
}
