package client

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"testing"

	"github.com/cyllective/oauth-labs/lab01/client/internal/dto"
)

func TestDecryptField(t *testing.T) {
	key := bytes.Repeat([]byte{9}, 32)
	aad := []byte("user-1|calendar:user-1|title|v1")
	nonce := bytes.Repeat([]byte{3}, 12)
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatal(err)
	}
	ciphertext := gcm.Seal(nil, nonce, []byte("Security review"), aad)

	plaintext, err := decryptField(
		dto.FieldKey{
			Algorithm: "AES-256-GCM",
			KeyID:     "test-key",
			Key:       base64.RawURLEncoding.EncodeToString(key),
			AAD:       base64.RawURLEncoding.EncodeToString(aad),
		},
		dto.EncryptedField{
			Algorithm:  "AES-256-GCM",
			KeyID:      "test-key",
			Nonce:      base64.RawURLEncoding.EncodeToString(nonce),
			Ciphertext: base64.RawURLEncoding.EncodeToString(ciphertext),
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if plaintext != "Security review" {
		t.Fatalf("unexpected plaintext %q", plaintext)
	}
}
