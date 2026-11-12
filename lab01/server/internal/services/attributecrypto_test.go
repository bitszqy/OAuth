package services

import (
	"bytes"
	"testing"
)

func TestAttributeCryptoRoundTripAndAADBinding(t *testing.T) {
	service := NewAttributeCryptoService(bytes.Repeat([]byte{0x42}, 32))
	field, err := service.Encrypt("user-1", "calendar:user-1", "title", "Security review")
	if err != nil {
		t.Fatal(err)
	}
	plaintext, err := service.Decrypt("user-1", "calendar:user-1", "title", field)
	if err != nil {
		t.Fatal(err)
	}
	if plaintext != "Security review" {
		t.Fatalf("unexpected plaintext %q", plaintext)
	}
	if _, err := service.Decrypt("user-1", "calendar:user-1", "location", field); err == nil {
		t.Fatal("expected AAD mismatch to fail")
	}
}

func TestAttributeCryptoRejectsTamperedCiphertext(t *testing.T) {
	service := NewAttributeCryptoService(bytes.Repeat([]byte{0x24}, 32))
	field, err := service.Encrypt("user-1", "calendar:user-1", "title", "private")
	if err != nil {
		t.Fatal(err)
	}
	field.Ciphertext = field.Ciphertext[:len(field.Ciphertext)-1] + "A"
	if _, err := service.Decrypt("user-1", "calendar:user-1", "title", field); err == nil {
		t.Fatal("expected tampered ciphertext to fail")
	}
}

func TestAttributeCryptoKeyIsContextBound(t *testing.T) {
	service := NewAttributeCryptoService(bytes.Repeat([]byte{0x11}, 32))
	first, err := service.Key("user-1", "calendar:user-1", "title")
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.Key("user-2", "calendar:user-2", "title")
	if err != nil {
		t.Fatal(err)
	}
	if first.Key == second.Key || first.KeyID == second.KeyID {
		t.Fatal("keys must differ across users and resources")
	}
}
