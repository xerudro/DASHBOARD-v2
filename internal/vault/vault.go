package vault

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var (
	ErrVaultLocked       = errors.New("vault is locked")
	ErrVaultUnlocked     = errors.New("vault is already unlocked")
	ErrInvalidMasterKey  = errors.New("invalid master key")
	ErrPermissionDenied  = errors.New("permission denied")
)

// VaultConfig holds vault configuration
type VaultConfig struct {
	MasterKeyEnvVar string        // Environment variable for master key
	AutoLock        bool           // Auto-lock vault after period of inactivity
	AutoLockTimeout time.Duration  // Timeout before auto-lock
	EnableAudit     bool           // Enable audit logging
	EncryptionConfig EncryptionConfig
}

// DefaultVaultConfig returns default vault configuration
func DefaultVaultConfig() VaultConfig {
	return VaultConfig{
		MasterKeyEnvVar: "VAULT_MASTER_KEY",
		AutoLock:        true,
		AutoLockTimeout: 15 * time.Minute,
		EnableAudit:     true,
		EncryptionConfig: DefaultEncryptionConfig(),
	}
}

// Vault is the main secrets management service
type Vault struct {
	config      VaultConfig
	encryption  *EncryptionService
	storage     *StorageBackend
	masterKey   string
	locked      bool
	lastAccess  time.Time
	mu          sync.RWMutex
	autoLockTimer *time.Timer
}

// NewVault creates a new vault instance
func NewVault(db *sqlx.DB, config VaultConfig) (*Vault, error) {
	storage, err := NewStorageBackend(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage backend: %w", err)
	}

	encryption := NewEncryptionService(config.EncryptionConfig)

	vault := &Vault{
		config:     config,
		encryption: encryption,
		storage:    storage,
		locked:     true,
		lastAccess: time.Now(),
	}

	return vault, nil
}

// Unlock unlocks the vault with the master key
func (v *Vault) Unlock(masterKey string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if !v.locked {
		return ErrVaultUnlocked
	}

	// Verify master key (in production, this should verify against stored hash)
	if masterKey == "" {
		return ErrInvalidMasterKey
	}

	v.masterKey = masterKey
	v.locked = false
	v.lastAccess = time.Now()

	// Start auto-lock timer if enabled
	if v.config.AutoLock {
		v.startAutoLockTimer()
	}

	return nil
}

// UnlockFromEnv unlocks the vault using master key from environment variable
func (v *Vault) UnlockFromEnv() error {
	masterKey := os.Getenv(v.config.MasterKeyEnvVar)
	if masterKey == "" {
		return fmt.Errorf("master key not found in environment variable: %s", v.config.MasterKeyEnvVar)
	}

	return v.Unlock(masterKey)
}

// Lock locks the vault
func (v *Vault) Lock() {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.masterKey = ""
	v.locked = true

	if v.autoLockTimer != nil {
		v.autoLockTimer.Stop()
	}
}

// IsLocked returns whether the vault is locked
func (v *Vault) IsLocked() bool {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.locked
}

// startAutoLockTimer starts the auto-lock timer
func (v *Vault) startAutoLockTimer() {
	if v.autoLockTimer != nil {
		v.autoLockTimer.Stop()
	}

	v.autoLockTimer = time.AfterFunc(v.config.AutoLockTimeout, func() {
		v.Lock()
	})
}

// resetAutoLockTimer resets the auto-lock timer
func (v *Vault) resetAutoLockTimer() {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.lastAccess = time.Now()
	if v.config.AutoLock && !v.locked {
		v.startAutoLockTimer()
	}
}

// checkLocked returns error if vault is locked
func (v *Vault) checkLocked() error {
	if v.IsLocked() {
		return ErrVaultLocked
	}
	v.resetAutoLockTimer()
	return nil
}

// CreateSecret creates a new encrypted secret
func (v *Vault) CreateSecret(ctx context.Context, path, value, description string, userID uuid.UUID, expiresIn *time.Duration) error {
	if err := v.checkLocked(); err != nil {
		return err
	}

	// Encrypt the secret value
	encrypted, err := v.encryption.Encrypt(value, v.masterKey)
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	// Prepare secret
	secret := &Secret{
		Path:        path,
		Value:       encrypted,
		Description: description,
		CreatedBy:   userID,
		Version:     1,
	}

	// Set expiration if provided
	if expiresIn != nil {
		expiresAt := time.Now().Add(*expiresIn)
		secret.ExpiresAt = &expiresAt
	}

	// Store secret
	err = v.storage.CreateSecret(ctx, secret)
	if err != nil {
		return err
	}

	// Audit log
	if v.config.EnableAudit {
		v.logAccess(ctx, secret.ID, userID, "create", "", true)
	}

	return nil
}

// GetSecret retrieves and decrypts a secret
func (v *Vault) GetSecret(ctx context.Context, path string, userID uuid.UUID, ipAddress string) (string, error) {
	if err := v.checkLocked(); err != nil {
		return "", err
	}

	// Retrieve encrypted secret
	secret, err := v.storage.GetSecret(ctx, path)
	if err != nil {
		if v.config.EnableAudit {
			v.logAccess(ctx, uuid.Nil, userID, "read", ipAddress, false)
		}
		return "", err
	}

	// Decrypt the secret value
	decrypted, err := v.encryption.Decrypt(secret.Value, v.masterKey)
	if err != nil {
		if v.config.EnableAudit {
			v.logAccess(ctx, secret.ID, userID, "read", ipAddress, false)
		}
		return "", fmt.Errorf("decryption failed: %w", err)
	}

	// Audit log
	if v.config.EnableAudit {
		v.logAccess(ctx, secret.ID, userID, "read", ipAddress, true)
	}

	return decrypted, nil
}

// UpdateSecret updates an existing secret
func (v *Vault) UpdateSecret(ctx context.Context, path, newValue string, userID uuid.UUID) error {
	if err := v.checkLocked(); err != nil {
		return err
	}

	// Get the secret first to get its ID for audit
	secret, err := v.storage.GetSecret(ctx, path)
	if err != nil {
		return err
	}

	// Encrypt the new value
	encrypted, err := v.encryption.Encrypt(newValue, v.masterKey)
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	// Update secret
	err = v.storage.UpdateSecret(ctx, path, encrypted, userID)
	if err != nil {
		return err
	}

	// Audit log
	if v.config.EnableAudit {
		v.logAccess(ctx, secret.ID, userID, "update", "", true)
	}

	return nil
}

// DeleteSecret deletes a secret
func (v *Vault) DeleteSecret(ctx context.Context, path string, userID uuid.UUID) error {
	if err := v.checkLocked(); err != nil {
		return err
	}

	// Get the secret first to get its ID for audit
	secret, err := v.storage.GetSecret(ctx, path)
	if err != nil {
		return err
	}

	// Delete secret
	err = v.storage.DeleteSecret(ctx, path)
	if err != nil {
		return err
	}

	// Audit log
	if v.config.EnableAudit {
		v.logAccess(ctx, secret.ID, userID, "delete", "", true)
	}

	return nil
}

// ListSecrets lists all secrets under a path prefix (without decrypted values)
func (v *Vault) ListSecrets(ctx context.Context, pathPrefix string) ([]Secret, error) {
	if err := v.checkLocked(); err != nil {
		return nil, err
	}

	return v.storage.ListSecrets(ctx, pathPrefix)
}

// RotateSecret rotates a secret with a new master key
func (v *Vault) RotateSecret(ctx context.Context, path, newMasterKey string, userID uuid.UUID) error {
	if err := v.checkLocked(); err != nil {
		return err
	}

	// Get current secret
	secret, err := v.storage.GetSecret(ctx, path)
	if err != nil {
		return err
	}

	// Re-encrypt with new master key
	newEncrypted, err := v.encryption.RotateEncryption(secret.Value, v.masterKey, newMasterKey)
	if err != nil {
		return fmt.Errorf("rotation failed: %w", err)
	}

	// Update secret
	err = v.storage.UpdateSecret(ctx, path, newEncrypted, userID)
	if err != nil {
		return err
	}

	// Audit log
	if v.config.EnableAudit {
		v.logAccess(ctx, secret.ID, userID, "rotate", "", true)
	}

	return nil
}

// GetSecretVersion retrieves a specific version of a secret
func (v *Vault) GetSecretVersion(ctx context.Context, path string, version int, userID uuid.UUID) (string, error) {
	if err := v.checkLocked(); err != nil {
		return "", err
	}

	// Retrieve encrypted secret version
	secretVersion, err := v.storage.GetSecretVersion(ctx, path, version)
	if err != nil {
		return "", err
	}

	// Decrypt the secret value
	decrypted, err := v.encryption.Decrypt(secretVersion.Value, v.masterKey)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %w", err)
	}

	// Audit log
	if v.config.EnableAudit {
		v.logAccess(ctx, secretVersion.SecretID, userID, "read_version", "", true)
	}

	return decrypted, nil
}

// ListSecretVersions lists all versions of a secret
func (v *Vault) ListSecretVersions(ctx context.Context, path string) ([]SecretVersion, error) {
	if err := v.checkLocked(); err != nil {
		return nil, err
	}

	return v.storage.ListSecretVersions(ctx, path)
}

// GetAuditLogs retrieves audit logs for a secret
func (v *Vault) GetAuditLogs(ctx context.Context, path string, limit int) ([]AuditLog, error) {
	if err := v.checkLocked(); err != nil {
		return nil, err
	}

	secret, err := v.storage.GetSecret(ctx, path)
	if err != nil {
		return nil, err
	}

	return v.storage.GetAuditLogs(ctx, secret.ID, limit)
}

// CleanupExpiredSecrets removes expired secrets
func (v *Vault) CleanupExpiredSecrets(ctx context.Context) (int64, error) {
	if err := v.checkLocked(); err != nil {
		return 0, err
	}

	return v.storage.CleanupExpiredSecrets(ctx)
}

// GenerateSecureToken generates a cryptographically secure token
func (v *Vault) GenerateSecureToken(length int) (string, error) {
	return v.encryption.GenerateSecureToken(length)
}

// logAccess logs an access event
func (v *Vault) logAccess(ctx context.Context, secretID, userID uuid.UUID, action, ipAddress string, success bool) {
	log := &AuditLog{
		SecretID:  secretID,
		UserID:    userID,
		Action:    action,
		IPAddress: ipAddress,
		Success:   success,
		Timestamp: time.Now(),
	}

	// Log asynchronously to not block operations
	go v.storage.LogAccess(context.Background(), log)
}

// RotateAllSecrets rotates all secrets with a new master key
// This is a maintenance operation for key rotation
func (v *Vault) RotateAllSecrets(ctx context.Context, newMasterKey string, userID uuid.UUID) error {
	if err := v.checkLocked(); err != nil {
		return err
	}

	// Get all secrets
	secrets, err := v.storage.ListSecrets(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to list secrets: %w", err)
	}

	// Rotate each secret
	for _, secret := range secrets {
		err := v.RotateSecret(ctx, secret.Path, newMasterKey, userID)
		if err != nil {
			return fmt.Errorf("failed to rotate secret %s: %w", secret.Path, err)
		}
	}

	// Update master key
	v.mu.Lock()
	v.masterKey = newMasterKey
	v.mu.Unlock()

	return nil
}

// Health checks the vault health status
func (v *Vault) Health() map[string]interface{} {
	v.mu.RLock()
	defer v.mu.RUnlock()

	return map[string]interface{}{
		"locked":      v.locked,
		"last_access": v.lastAccess,
		"auto_lock":   v.config.AutoLock,
		"audit":       v.config.EnableAudit,
	}
}
