package vault

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
)

// EncryptionConfig holds encryption configuration
type EncryptionConfig struct {
	// Argon2 parameters for key derivation
	Argon2Time    uint32 // Number of iterations (recommended: 3-4)
	Argon2Memory  uint32 // Memory in KiB (recommended: 64MB = 65536)
	Argon2Threads uint8  // Number of threads (recommended: 4)
	Argon2KeyLen  uint32 // Key length in bytes (32 for AES-256)
}

// DefaultEncryptionConfig returns secure default encryption settings
func DefaultEncryptionConfig() EncryptionConfig {
	return EncryptionConfig{
		Argon2Time:    3,
		Argon2Memory:  65536, // 64 MB
		Argon2Threads: 4,
		Argon2KeyLen:  32, // 256 bits for AES-256
	}
}

// EncryptionService handles all encryption/decryption operations
type EncryptionService struct {
	config EncryptionConfig
}

// NewEncryptionService creates a new encryption service
func NewEncryptionService(config EncryptionConfig) *EncryptionService {
	return &EncryptionService{
		config: config,
	}
}

// DeriveKey derives an encryption key from a password using Argon2
// Returns: derived key (32 bytes), salt (16 bytes), error
func (e *EncryptionService) DeriveKey(password string, salt []byte) ([]byte, []byte, error) {
	// Generate new salt if not provided
	if salt == nil {
		salt = make([]byte, 16)
		if _, err := io.ReadFull(rand.Reader, salt); err != nil {
			return nil, nil, fmt.Errorf("failed to generate salt: %w", err)
		}
	}

	// Derive key using Argon2id (most secure variant)
	key := argon2.IDKey(
		[]byte(password),
		salt,
		e.config.Argon2Time,
		e.config.Argon2Memory,
		e.config.Argon2Threads,
		e.config.Argon2KeyLen,
	)

	return key, salt, nil
}

// Encrypt encrypts plaintext using AES-256-GCM
// Format: [salt(16)][nonce(12)][ciphertext][tag(16)]
func (e *EncryptionService) Encrypt(plaintext, password string) (string, error) {
	if plaintext == "" {
		return "", errors.New("plaintext cannot be empty")
	}
	if password == "" {
		return "", errors.New("password cannot be empty")
	}

	// Derive encryption key
	key, salt, err := e.DeriveKey(password, nil)
	if err != nil {
		return "", fmt.Errorf("key derivation failed: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode (Galois/Counter Mode - provides authenticated encryption)
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate nonce (12 bytes for GCM)
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt and authenticate
	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)

	// Combine: salt + nonce + ciphertext (includes authentication tag)
	combined := make([]byte, len(salt)+len(nonce)+len(ciphertext))
	copy(combined[0:], salt)
	copy(combined[len(salt):], nonce)
	copy(combined[len(salt)+len(nonce):], ciphertext)

	// Encode to base64 for safe storage
	encoded := base64.StdEncoding.EncodeToString(combined)
	return encoded, nil
}

// Decrypt decrypts ciphertext using AES-256-GCM
func (e *EncryptionService) Decrypt(encoded, password string) (string, error) {
	if encoded == "" {
		return "", errors.New("ciphertext cannot be empty")
	}
	if password == "" {
		return "", errors.New("password cannot be empty")
	}

	// Decode from base64
	combined, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	// Minimum length: salt(16) + nonce(12) + tag(16) = 44 bytes
	if len(combined) < 44 {
		return "", errors.New("ciphertext too short")
	}

	// Extract components
	salt := combined[0:16]
	nonce := combined[16:28]
	ciphertext := combined[28:]

	// Derive decryption key using stored salt
	key, _, err := e.DeriveKey(password, salt)
	if err != nil {
		return "", fmt.Errorf("key derivation failed: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Decrypt and verify authentication tag
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decryption failed (wrong password or corrupted data): %w", err)
	}

	return string(plaintext), nil
}

// HashPassword creates a SHA-256 hash of a password (for storage/comparison)
// Note: This is different from DeriveKey. Use this for password verification,
// not for encryption keys.
func (e *EncryptionService) HashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return base64.StdEncoding.EncodeToString(hash[:])
}

// GenerateRandomKey generates a cryptographically secure random key
func (e *EncryptionService) GenerateRandomKey(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("key length must be positive")
	}

	key := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return "", fmt.Errorf("failed to generate random key: %w", err)
	}

	return base64.StdEncoding.EncodeToString(key), nil
}

// GenerateSecureToken generates a secure random token (useful for API keys)
func (e *EncryptionService) GenerateSecureToken(length int) (string, error) {
	return e.GenerateRandomKey(length)
}

// VerifyIntegrity verifies the integrity of encrypted data
// Returns true if data can be decrypted (implies integrity is valid)
func (e *EncryptionService) VerifyIntegrity(encoded, password string) bool {
	_, err := e.Decrypt(encoded, password)
	return err == nil
}

// RotateEncryption re-encrypts data with a new password
// This is useful for key rotation
func (e *EncryptionService) RotateEncryption(encoded, oldPassword, newPassword string) (string, error) {
	// Decrypt with old password
	plaintext, err := e.Decrypt(encoded, oldPassword)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt with old password: %w", err)
	}

	// Re-encrypt with new password
	newEncoded, err := e.Encrypt(plaintext, newPassword)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt with new password: %w", err)
	}

	return newEncoded, nil
}
