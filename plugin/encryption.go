package replicator

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/openbao/openbao/sdk/v2/logical"
)

const (
	encryptionKeyStoragePath = "internal/encryption_key"
)

type Encrypter struct {
	backend *Backend
}

func NewEncrypter(b *Backend) *Encrypter {
	return &Encrypter{backend: b}
}

func (e *Encrypter) Encrypt(plaintext string) (string, error) {
	key, err := e.getOrCreateKey()
	if err != nil {
		return "", fmt.Errorf("failed to get encryption key: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (e *Encrypter) Decrypt(encrypted string) (string, error) {
	key, err := e.getOrCreateKey()
	if err != nil {
		return "", fmt.Errorf("failed to get encryption key: %w", err)
	}

	data, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

func (e *Encrypter) getOrCreateKey() ([]byte, error) {
	ctx := context.Background()
	storage := e.backend.storage

	entry, err := storage.Get(ctx, encryptionKeyStoragePath)
	if err != nil {
		return nil, err
	}

	if entry != nil {
		keyData := entry.Value
		if len(keyData) == 0 {
			return nil, fmt.Errorf("empty key in storage")
		}
		return keyData, nil
	}

	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	entry = &logical.StorageEntry{
		Key:   encryptionKeyStoragePath,
		Value: key,
	}
	if err := storage.Put(ctx, entry); err != nil {
		return nil, fmt.Errorf("failed to store key: %w", err)
	}

	return key, nil
}

type SecureConfig struct {
	VaultAddress     string `json:"vault_address"`
	VaultMount       string `json:"vault_mount"`
	AppRoleRoleID    string `json:"approle_role_id"`
	AppRoleSecretID  string `json:"approle_secret_id"`
	DestinationToken string `json:"destination_token"`
	DestinationMount string `json:"destination_mount"`
}

func (b *Backend) encryptConfig(ctx context.Context, config *Configuration) (*SecureConfig, error) {
	encrypter := NewEncrypter(b)

	encryptedRoleID, err := encrypter.Encrypt(config.AppRoleRoleID)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt role_id: %w", err)
	}

	encryptedSecretID, err := encrypter.Encrypt(config.AppRoleSecretID)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt secret_id: %w", err)
	}

	encryptedToken, err := encrypter.Encrypt(config.DestinationToken)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt token: %w", err)
	}

	return &SecureConfig{
		VaultAddress:     config.VaultAddress,
		VaultMount:       config.VaultMount,
		AppRoleRoleID:    encryptedRoleID,
		AppRoleSecretID:  encryptedSecretID,
		DestinationToken: encryptedToken,
		DestinationMount: config.DestinationMount,
	}, nil
}

func (b *Backend) decryptConfig(ctx context.Context, secureConfig *SecureConfig) (*Configuration, error) {
	encrypter := NewEncrypter(b)

	decryptedRoleID, err := encrypter.Decrypt(secureConfig.AppRoleRoleID)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt role_id: %w", err)
	}

	decryptedSecretID, err := encrypter.Decrypt(secureConfig.AppRoleSecretID)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt secret_id: %w", err)
	}

	decryptedToken, err := encrypter.Decrypt(secureConfig.DestinationToken)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt token: %w", err)
	}

	return &Configuration{
		VaultAddress:     secureConfig.VaultAddress,
		VaultMount:       secureConfig.VaultMount,
		AppRoleRoleID:    decryptedRoleID,
		AppRoleSecretID:  decryptedSecretID,
		DestinationToken: decryptedToken,
		DestinationMount: secureConfig.DestinationMount,
	}, nil
}

func (b *Backend) writeEncryptedConfig(ctx context.Context, storage logical.Storage, config *Configuration) error {
	secureConfig, err := b.encryptConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to encrypt config: %w", err)
	}

	entry, err := logical.StorageEntryJSON(configStoragePath, secureConfig)
	if err != nil {
		return err
	}

	return storage.Put(ctx, entry)
}

func (b *Backend) readEncryptedConfig(ctx context.Context, storage logical.Storage) (*Configuration, error) {
	entry, err := storage.Get(ctx, configStoragePath)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, nil
	}

	var secureConfig SecureConfig
	if err := entry.DecodeJSON(&secureConfig); err != nil {
		return nil, err
	}

	return b.decryptConfig(ctx, &secureConfig)
}
