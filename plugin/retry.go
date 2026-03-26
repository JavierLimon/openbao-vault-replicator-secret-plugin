package replicator

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/hashicorp/vault/api"
)

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxRetries      int
	InitialInterval time.Duration
	MaxInterval     time.Duration
	Multiplier      float64
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:      3,
		InitialInterval: 500 * time.Millisecond,
		MaxInterval:     10 * time.Second,
		Multiplier:      2.0,
	}
}

// RetryableOperation defines an operation that can be retried
type RetryableOperation func() error

// RetryWithBackoff executes an operation with exponential backoff retry
func RetryWithBackoff(ctx context.Context, config *RetryConfig, operation RetryableOperation) error {
	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = config.InitialInterval
	bo.MaxInterval = config.MaxInterval
	bo.Multiplier = config.Multiplier
	bo.MaxElapsedTime = 0 // No max time limit, use max retries

	var lastError error
	retryCount := 0

	for {
		err := operation()
		if err == nil {
			return nil
		}

		retryCount++
		if retryCount > config.MaxRetries {
			return fmt.Errorf("max retries (%d) exceeded, last error: %w", config.MaxRetries, lastError)
		}

		lastError = err

		// Check if error is retryable
		if !isRetryableError(err) {
			return err
		}

		// Check context cancellation
		select {
		case <-ctx.Done():
			return fmt.Errorf("operation canceled: %w", ctx.Err())
		default:
		}

		// Wait before retry
		nextBackOff := bo.NextBackOff()
		if nextBackOff == backoff.Stop {
			return fmt.Errorf("backoff stopped, last error: %w", lastError)
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("operation canceled during backoff: %w", ctx.Err())
		case <-time.After(nextBackOff):
			// Continue to retry
		}
	}
}

// isRetryableError determines if an error should trigger a retry
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())

	// Network-related errors are retryable
	retryablePatterns := []string{
		"connection refused",
		"connection reset",
		"connection timed out",
		"timeout",
		"no such host",
		"network is unreachable",
		"i/o timeout",
		"temporary failure",
		"server error",
		"service unavailable",
		"503",
		"502",
		"429", // rate limited
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	return false
}

// ValidateConfig validates the plugin configuration
func ValidateConfig(config *Configuration) error {
	if config == nil {
		return fmt.Errorf("configuration is nil")
	}

	// Validate Vault address
	if strings.TrimSpace(config.VaultAddress) == "" {
		return fmt.Errorf("vault_address is required")
	}

	if err := validateURL(config.VaultAddress, "vault_address"); err != nil {
		return err
	}

	// Validate Vault mount
	if strings.TrimSpace(config.VaultMount) == "" {
		return fmt.Errorf("vault_mount is required")
	}

	if err := validateMountPath(config.VaultMount); err != nil {
		return err
	}

	// Validate AppRole role_id
	if strings.TrimSpace(config.AppRoleRoleID) == "" {
		return fmt.Errorf("approle_role_id is required")
	}

	// Validate AppRole secret_id
	if strings.TrimSpace(config.AppRoleSecretID) == "" {
		return fmt.Errorf("approle_secret_id is required")
	}

	// Validate destination token
	if strings.TrimSpace(config.DestinationToken) == "" {
		return fmt.Errorf("destination_token is required")
	}

	// Validate destination mount
	if strings.TrimSpace(config.DestinationMount) == "" {
		return fmt.Errorf("destination_mount is required")
	}

	if err := validateMountPath(config.DestinationMount); err != nil {
		return err
	}

	return nil
}

// validateURL validates that a string is a valid URL
func validateURL(value, fieldName string) error {
	u, err := url.Parse(value)
	if err != nil {
		return fmt.Errorf("%s is not a valid URL: %w", fieldName, err)
	}

	if u.Scheme == "" {
		return fmt.Errorf("%s must include a scheme (http or https)", fieldName)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("%s must use http or https scheme", fieldName)
	}

	if u.Host == "" {
		return fmt.Errorf("%s must include a host", fieldName)
	}

	return nil
}

// validateMountPath validates a KV mount path
func validateMountPath(path string) error {
	// Mount paths should not contain leading slash for internal operations
	path = strings.TrimPrefix(path, "/")

	if strings.Contains(path, "..") {
		return fmt.Errorf("mount path must not contain '..'")
	}

	// Check for invalid characters
	invalidChars := []string{" ", "\t", "\n", "\r", "\x00"}
	for _, char := range invalidChars {
		if strings.Contains(path, char) {
			return fmt.Errorf("mount path contains invalid characters")
		}
	}

	return nil
}

// NewVaultClient creates a new Vault client with AppRole authentication
func NewVaultClient(addr, mount, roleID, secretID string) (*api.Client, error) {
	config := api.DefaultConfig()
	config.Address = addr

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create vault client: %w", err)
	}

	return client, nil
}

// LoginToVault performs AppRole login and returns the client token
func LoginToVault(client *api.Client, roleID, secretID string) (string, error) {
	resp, err := client.Logical().Write("auth/approle/login", map[string]interface{}{
		"role_id":   roleID,
		"secret_id": secretID,
	})
	if err != nil {
		return "", fmt.Errorf("approle login failed: %w", err)
	}

	if resp == nil {
		return "", fmt.Errorf("approle login returned nil response")
	}

	// Try resp.Data["client_token"] first (standard Vault response)
	if tokenRaw, ok := resp.Data["client_token"]; ok {
		if token, ok := tokenRaw.(string); ok && token != "" {
			return token, nil
		}
	}

	// Try resp.Data["token"] as fallback
	if tokenRaw, ok := resp.Data["token"]; ok {
		if token, ok := tokenRaw.(string); ok && token != "" {
			return token, nil
		}
	}

	// Try resp.Auth.ClientToken (for HashiCorp Vault responses wrapped in auth)
	if resp.Auth != nil && resp.Auth.ClientToken != "" {
		return resp.Auth.ClientToken, nil
	}

	return "", fmt.Errorf("token not found in approle login response - response: %+v", resp)
}

// ListOrganizationsWithRetry lists organizations with retry logic
func ListOrganizationsWithRetry(ctx context.Context, client *api.Client, mount, orgPath string) ([]string, error) {
	var result []string
	err := RetryWithBackoff(ctx, DefaultRetryConfig(), func() error {
		orgs, err := listOrganizationsInternal(client, mount, orgPath)
		if err != nil {
			return err
		}
		result = orgs
		return nil
	})
	return result, err
}

// ListSecretsInOrgWithRetry lists secrets in an organization with retry logic
func ListSecretsInOrgWithRetry(ctx context.Context, client *api.Client, mount, orgPath, org string) ([]string, error) {
	var result []string
	err := RetryWithBackoff(ctx, DefaultRetryConfig(), func() error {
		secrets, err := listSecretsInOrgInternal(client, mount, org)
		if err != nil {
			return err
		}
		result = secrets
		return nil
	})
	return result, err
}

// ReadSecretWithRetry reads a secret with retry logic
func ReadSecretWithRetry(ctx context.Context, client *api.Client, mount, orgPath, org, secret string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := RetryWithBackoff(ctx, DefaultRetryConfig(), func() error {
		data, err := readSecretInternal(client, mount, org, secret)
		if err != nil {
			return err
		}
		result = data
		return nil
	})
	return result, err
}

// listOrganizationsInternal lists organizations (internal implementation)
func listOrganizationsInternal(client *api.Client, mount, orgPath string) ([]string, error) {
	orgPath = strings.TrimSuffix(orgPath, "/")
	if orgPath == "" || orgPath == "." {
		orgPath = ""
	} else {
		orgPath = orgPath + "/"
	}
	path := fmt.Sprintf("%s/metadata/%s", mount, orgPath)
	resp, err := client.Logical().List(path)
	if err != nil {
		return nil, fmt.Errorf("failed to list organizations at %s: %w", path, err)
	}
	if resp == nil {
		return []string{}, nil
	}

	keysRaw, ok := resp.Data["keys"]
	if !ok {
		return []string{}, nil
	}

	keys, ok := keysRaw.([]interface{})
	if !ok {
		return []string{}, nil
	}

	var orgs []string
	for _, k := range keys {
		if keyStr, ok := k.(string); ok {
			orgs = append(orgs, keyStr)
		}
	}

	return orgs, nil
}

// listSecretsInOrgInternal lists secrets in an organization (internal implementation)
func listSecretsInOrgInternal(client *api.Client, mount, org string) ([]string, error) {
	org = strings.TrimSuffix(org, "/")
	org = strings.ReplaceAll(org, "//", "/")
	return listSecretsRecursive(client, mount, org)
}

func listSecretsRecursive(client *api.Client, mount, currentPath string) ([]string, error) {
	path := fmt.Sprintf("%s/metadata/%s/", mount, currentPath)
	resp, err := client.Logical().List(path)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets at %s: %w", currentPath, err)
	}
	if resp == nil {
		return []string{}, nil
	}

	keysRaw, ok := resp.Data["keys"]
	if !ok {
		return []string{}, nil
	}

	keys, ok := keysRaw.([]interface{})
	if !ok {
		return []string{}, nil
	}

	var secrets []string
	for _, k := range keys {
		if keyStr, ok := k.(string); ok {
			if strings.HasSuffix(keyStr, "/") {
				folderName := strings.TrimSuffix(keyStr, "/")
				subPath := currentPath + "/" + folderName
				subSecrets, err := listSecretsRecursive(client, mount, subPath)
				if err != nil {
					return nil, err
				}
				secrets = append(secrets, subSecrets...)
			} else {
				secrets = append(secrets, keyStr)
			}
		}
	}

	return secrets, nil
}

// readSecretInternal reads a secret (internal implementation)
func readSecretInternal(client *api.Client, mount, org, secret string) (map[string]interface{}, error) {
	org = strings.TrimSuffix(org, "/")
	org = strings.ReplaceAll(org, "//", "/")
	path := fmt.Sprintf("%s/data/%s/%s", mount, org, secret)
	resp, err := client.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret %s/%s: %w", org, secret, err)
	}
	if resp == nil || resp.Data == nil {
		return nil, nil
	}

	dataRaw, ok := resp.Data["data"]
	if !ok {
		return nil, nil
	}

	data, ok := dataRaw.(map[string]interface{})
	if !ok {
		return nil, nil
	}

	return data, nil
}

type SecretData struct {
	Data           map[string]interface{} `json:"data"`
	CustomMetadata map[string]interface{} `json:"custom_metadata,omitempty"`
	Version        int                    `json:"version"`
}

func ReadSecretWithMetadata(ctx context.Context, client *api.Client, mount, orgPath, org, secret string) (*SecretData, error) {
	var result *SecretData
	err := RetryWithBackoff(ctx, DefaultRetryConfig(), func() error {
		data, err := readSecretWithMetadataInternal(client, mount, org, secret)
		if err != nil {
			return err
		}
		result = data
		return nil
	})
	return result, err
}

func readSecretWithMetadataInternal(client *api.Client, mount, org, secret string) (*SecretData, error) {
	org = strings.TrimSuffix(org, "/")
	org = strings.ReplaceAll(org, "//", "/")

	dataPath := fmt.Sprintf("%s/data/%s/%s", mount, org, secret)
	dataResp, err := client.Logical().Read(dataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret %s/%s: %w", org, secret, err)
	}
	if dataResp == nil || dataResp.Data == nil {
		return nil, nil
	}

	dataRaw, ok := dataResp.Data["data"]
	if !ok {
		return nil, nil
	}
	data, ok := dataRaw.(map[string]interface{})
	if !ok {
		return nil, nil
	}

	version := 1
	if v, ok := dataResp.Data["version"].(float64); ok {
		version = int(v)
	}

	metadataPath := fmt.Sprintf("%s/metadata/%s/%s", mount, org, secret)
	metadataResp, err := client.Logical().Read(metadataPath)
	customMetadata := make(map[string]interface{})
	if err == nil && metadataResp != nil && metadataResp.Data != nil {
		if cm, ok := metadataResp.Data["custom_metadata"].(map[string]interface{}); ok {
			customMetadata = cm
		}
	}

	return &SecretData{
		Data:           data,
		CustomMetadata: customMetadata,
		Version:        version,
	}, nil
}

const (
	defaultPageSize = 100
	maxPageSize     = 500
)

type ListOptions struct {
	Prefix   string
	PageSize int
}

func DefaultListOptions() *ListOptions {
	return &ListOptions{
		PageSize: defaultPageSize,
	}
}

func (opts *ListOptions) WithPageSize(pageSize int) *ListOptions {
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	opts.PageSize = pageSize
	return opts
}

func ListOrganizationsPaged(ctx context.Context, client *api.Client, mount, orgPath string, opts *ListOptions) ([]string, error) {
	if opts == nil {
		opts = DefaultListOptions()
	}

	var allOrgs []string
	marker := ""

	for {
		select {
		case <-ctx.Done():
			return allOrgs, ctx.Err()
		default:
		}

		path := fmt.Sprintf("%s/metadata/%s?list=true", mount, orgPath)
		if marker != "" {
			path += "&marker=" + marker
		}
		if opts.PageSize != defaultPageSize {
			path += fmt.Sprintf("&limit=%d", opts.PageSize)
		}

		var resp *api.Secret
		err := RetryWithBackoff(ctx, DefaultRetryConfig(), func() error {
			var err error
			resp, err = client.Logical().List(path)
			return err
		})
		if err != nil {
			return allOrgs, fmt.Errorf("failed to list organizations: %w", err)
		}

		if resp == nil {
			break
		}

		keysRaw, ok := resp.Data["keys"]
		if !ok {
			break
		}

		keys, ok := keysRaw.([]interface{})
		if !ok {
			break
		}

		for _, k := range keys {
			if keyStr, ok := k.(string); ok {
				allOrgs = append(allOrgs, keyStr)
			}
		}

		if len(keys) < opts.PageSize {
			break
		}
	}

	return allOrgs, nil
}

func ListSecretsInOrgPaged(ctx context.Context, client *api.Client, mount, orgPath, org string) ([]string, error) {
	var allSecrets []string

	path := fmt.Sprintf("%s/metadata/%s%s", mount, orgPath, org)
	resp, err := client.Logical().List(path)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets in org %s: %w", org, err)
	}

	if resp == nil {
		return []string{}, nil
	}

	keysRaw, ok := resp.Data["keys"]
	if !ok {
		return []string{}, nil
	}

	keys, ok := keysRaw.([]interface{})
	if !ok {
		return []string{}, nil
	}

	for _, k := range keys {
		if keyStr, ok := k.(string); ok {
			allSecrets = append(allSecrets, keyStr)
		}
	}

	return allSecrets, nil
}
