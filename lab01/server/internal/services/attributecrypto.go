package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/hkdf"

	"github.com/cyllective/oauth-labs/lab01/server/internal/dto"
)

const attributeEncryptionAlgorithm = "AES-256-GCM"

type AttributeCryptoService struct {
	masterKey []byte
}

func NewAttributeCryptoService(masterKey []byte) *AttributeCryptoService {
	return &AttributeCryptoService{masterKey: append([]byte(nil), masterKey...)}
}

func (s *AttributeCryptoService) Encrypt(userID, resourceID, attribute, value string) (dto.EncryptedField, error) {
	key, err := s.deriveKey(userID, resourceID, attribute)
	if err != nil {
		return dto.EncryptedField{}, err
	}
	aad := s.additionalData(userID, resourceID, attribute)
	block, err := aes.NewCipher(key)
	if err != nil {
		return dto.EncryptedField{}, fmt.Errorf("create field cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return dto.EncryptedField{}, fmt.Errorf("create field GCM: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return dto.EncryptedField{}, fmt.Errorf("generate field nonce: %w", err)
	}
	ciphertext := gcm.Seal(nil, nonce, []byte(value), aad)
	return dto.EncryptedField{
		Algorithm:  attributeEncryptionAlgorithm,
		KeyID:      s.keyID(userID, resourceID, attribute),
		Nonce:      base64.RawURLEncoding.EncodeToString(nonce),
		Ciphertext: base64.RawURLEncoding.EncodeToString(ciphertext),
	}, nil
}

func (s *AttributeCryptoService) Key(userID, resourceID, attribute string) (dto.FieldKey, error) {
	key, err := s.deriveKey(userID, resourceID, attribute)
	if err != nil {
		return dto.FieldKey{}, err
	}
	return dto.FieldKey{
		Algorithm: attributeEncryptionAlgorithm,
		KeyID:     s.keyID(userID, resourceID, attribute),
		Key:       base64.RawURLEncoding.EncodeToString(key),
		AAD:       base64.RawURLEncoding.EncodeToString(s.additionalData(userID, resourceID, attribute)),
	}, nil
}

func (s *AttributeCryptoService) Decrypt(userID, resourceID, attribute string, field dto.EncryptedField) (string, error) {
	if field.Algorithm != attributeEncryptionAlgorithm {
		return "", errors.New("unsupported attribute encryption algorithm")
	}
	key, err := s.deriveKey(userID, resourceID, attribute)
	if err != nil {
		return "", err
	}
	nonce, err := base64.RawURLEncoding.DecodeString(field.Nonce)
	if err != nil {
		return "", fmt.Errorf("decode field nonce: %w", err)
	}
	ciphertext, err := base64.RawURLEncoding.DecodeString(field.Ciphertext)
	if err != nil {
		return "", fmt.Errorf("decode field ciphertext: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	plaintext, err := gcm.Open(nil, nonce, ciphertext, s.additionalData(userID, resourceID, attribute))
	if err != nil {
		return "", fmt.Errorf("decrypt field: %w", err)
	}
	return string(plaintext), nil
}

func (s *AttributeCryptoService) deriveKey(userID, resourceID, attribute string) ([]byte, error) {
	if len(s.masterKey) != 32 {
		return nil, errors.New("attribute master key must be 32 bytes")
	}
	context := []byte(fmt.Sprintf("%s|%s|%s|v1", userID, resourceID, attribute))
	reader := hkdf.New(sha256.New, s.masterKey, []byte("oauth-field-encryption-v1"), context)
	key := make([]byte, 32)
	if _, err := io.ReadFull(reader, key); err != nil {
		return nil, fmt.Errorf("derive attribute key: %w", err)
	}
	return key, nil
}

func (s *AttributeCryptoService) keyID(userID, resourceID, attribute string) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%s|v1", userID, resourceID, attribute)))
	return base64.RawURLEncoding.EncodeToString(sum[:16])
}

func (s *AttributeCryptoService) additionalData(userID, resourceID, attribute string) []byte {
	return []byte(fmt.Sprintf("%s|%s|%s|v1", userID, resourceID, attribute))
}
