package replicator

import (
	"context"
	"fmt"
	"time"

	"github.com/openbao/openbao/sdk/v2/framework"
	"github.com/openbao/openbao/sdk/v2/logical"
)

const (
	syncStatusPath     = "sync/status"
	syncHistoryPrefix  = "sync/history/"
	syncHistoryListKey = "sync/history"
)

type SyncStatus struct {
	LastSync            time.Time `json:"last_sync"`
	CompletedAt         time.Time `json:"completed_at,omitempty"`
	StartedAt           time.Time `json:"started_at,omitempty"`
	Status              string    `json:"status"`
	LastError           string    `json:"last_error,omitempty"`
	LastOrg             string    `json:"last_org"`
	TotalSecrets        int       `json:"total_secrets"`
	SyncedSecrets       int       `json:"synced_secrets"`
	SyncedOrgs          int       `json:"synced_organizations"`
	TotalOrgs           int       `json:"total_organizations"`
	Failed              int       `json:"failed"`
	OrganizationsSynced int       `json:"organizations_synced,omitempty"`
	SecretsSynced       int       `json:"secrets_synced,omitempty"`
	DurationSeconds     int       `json:"duration_seconds,omitempty"`
}

type SyncHistoryEntry struct {
	Timestamp           time.Time `json:"timestamp"`
	Status              string    `json:"status"`
	OrganizationsSynced int       `json:"organizations_synced"`
	SecretsSynced       int       `json:"secrets_synced"`
	Failed              int       `json:"failed"`
	DurationSeconds     int       `json:"duration_seconds"`
}

func (b *Backend) pathSync() *framework.Path {
	return &framework.Path{
		Pattern: "sync/secrets",
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.CreateOperation: &framework.PathOperation{
				Summary:     "Trigger secret replication",
				Description: "Triggers the replication of secrets from HashiCorp Vault to OpenBao",
				Callback:    b.pathSyncSecrets,
			},
			logical.UpdateOperation: &framework.PathOperation{
				Summary:     "Trigger secret replication",
				Description: "Triggers the replication of secrets from HashiCorp Vault to OpenBao",
				Callback:    b.pathSyncSecrets,
			},
		},
		Fields: map[string]*framework.FieldSchema{
			"organizations": {
				Type:        framework.TypeStringSlice,
				Description: "Specific organizations to sync (optional)",
			},
			"dry_run": {
				Type:        framework.TypeBool,
				Description: "Preview only, do not write",
			},
		},
		HelpSynopsis:    "Trigger secret replication from Vault to OpenBao",
		HelpDescription: "Triggers the replication of secrets from HashiCorp Vault to OpenBao",
	}
}

func (b *Backend) pathSyncStatus() *framework.Path {
	return &framework.Path{
		Pattern: "sync/status",
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ReadOperation: &framework.PathOperation{
				Summary:     "Get current sync status",
				Description: "Returns the current status of the secret replication process",
				Callback:    b.pathSyncStatusRead,
			},
		},
		HelpSynopsis:    "Get the current sync status",
		HelpDescription: "Returns the current status of the secret replication process",
	}
}

func (b *Backend) pathSyncHistory() *framework.Path {
	return &framework.Path{
		Pattern: "sync/history",
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ListOperation: &framework.PathOperation{
				Summary:     "List sync history",
				Description: "Lists past sync operations",
				Callback:    b.pathSyncHistoryList,
			},
			logical.ReadOperation: &framework.PathOperation{
				Summary:     "Get sync history entry",
				Description: "Retrieves a specific sync history entry",
				Callback:    b.pathSyncHistoryRead,
			},
		},
		HelpSynopsis:    "List or get sync history",
		HelpDescription: "Lists past sync operations or retrieves a specific sync history entry",
	}
}

func (b *Backend) pathSyncHistoryTimestamp() *framework.Path {
	return &framework.Path{
		Pattern: "sync/history/" + framework.GenericNameRegex("timestamp"),
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ReadOperation: &framework.PathOperation{
				Summary:     "Get sync history entry by timestamp",
				Description: "Retrieves a specific sync history entry by its timestamp",
				Callback:    b.pathSyncHistoryTimestampRead,
			},
		},
		Fields: map[string]*framework.FieldSchema{
			"timestamp": {
				Type:        framework.TypeString,
				Description: "The timestamp of the sync history entry",
			},
		},
		HelpSynopsis:    "Get sync history entry by timestamp",
		HelpDescription: "Retrieves a specific sync history entry by its timestamp",
	}
}

func (b *Backend) pathSyncSecrets(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	config, err := b.readConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return logical.ErrorResponse("configuration not found"), logical.ErrInvalidRequest
	}

	// Validate configuration
	if err := ValidateConfig(config); err != nil {
		return logical.ErrorResponse("invalid configuration: " + err.Error()), logical.ErrInvalidRequest
	}

	orgsRaw, ok := data.Get("organizations").([]string)
	if !ok {
		return logical.ErrorResponse("organizations must be a string slice"), logical.ErrInvalidRequest
	}
	orgs := orgsRaw
	dryRunRaw, ok := data.Get("dry_run").(bool)
	if !ok {
		return logical.ErrorResponse("dry_run must be a boolean"), logical.ErrInvalidRequest
	}
	dryRun := dryRunRaw

	vaultClient, err := NewVaultClient(config.VaultAddress, config.VaultMount, config.AppRoleRoleID, config.AppRoleSecretID)
	if err != nil {
		return logical.ErrorResponse("failed to create Vault client: " + err.Error()), logical.ErrInvalidRequest
	}

	clientToken, err := LoginToVault(vaultClient, config.AppRoleRoleID, config.AppRoleSecretID)
	if err != nil {
		return logical.ErrorResponse("failed to login to Vault: " + err.Error()), logical.ErrInvalidRequest
	}
	defer func() {
		if clientToken != "" {
			if revokeErr := vaultClient.Auth().Token().RevokeSelf(clientToken); revokeErr != nil {
				b.logger.Warn("failed to revoke token", "error", revokeErr)
			}
		}
	}()

	vaultClient.SetToken(clientToken)

	status := &SyncStatus{
		Status:    "running",
		StartedAt: time.Now().UTC(),
	}

	if err = b.saveSyncStatus(ctx, req.Storage, status); err != nil {
		return nil, err
	}

	var allOrgs []string
	if len(orgs) > 0 {
		allOrgs = orgs
	} else {
		allOrgs, err = ListOrganizationsWithRetry(ctx, vaultClient, config.VaultMount, "")
		if err != nil {
			status.Status = "failed"
			status.LastError = err.Error()
			if saveErr := b.saveSyncStatus(ctx, req.Storage, status); saveErr != nil {
				b.logger.Error("failed to save sync status", "error", saveErr)
			}
			return logical.ErrorResponse("failed to list organizations: " + err.Error()), logical.ErrInvalidRequest
		}
	}

	var filteredOrgs []string
	for _, org := range allOrgs {
		if config.ShouldSyncOrg(org) {
			filteredOrgs = append(filteredOrgs, org)
		}
	}
	allOrgs = filteredOrgs

	status.TotalOrgs = len(allOrgs)
	status.LastOrg = ""

	for _, org := range allOrgs {
		status.LastOrg = org

		secrets, err := ListSecretsInOrgWithRetry(ctx, vaultClient, config.VaultMount, "", org)
		if err != nil {
			status.Failed++
			continue
		}

		status.TotalSecrets += len(secrets)

		for _, secret := range secrets {
			secretData, err := ReadSecretWithMetadata(ctx, vaultClient, config.VaultMount, "", org, secret)
			if err != nil {
				status.Failed++
				continue
			}
			if secretData == nil {
				status.Failed++
				continue
			}

			if !dryRun {
				if err := b.writeToLocalKVWithMetadata(org, secret, secretData.Data, secretData.CustomMetadata); err != nil {
					status.Failed++
					status.LastError = fmt.Sprintf("failed to write %s/%s: %v", org, secret, err)
					continue
				}
			}

			status.SyncedSecrets++
		}

		status.SyncedOrgs++
		if err := b.saveSyncStatus(ctx, req.Storage, status); err != nil {
			b.logger.Error("failed to save sync status", "error", err)
		}
	}

	status.Status = "completed"
	status.CompletedAt = time.Now().UTC()
	status.OrganizationsSynced = status.SyncedOrgs
	status.SecretsSynced = status.SyncedSecrets
	status.DurationSeconds = int(status.CompletedAt.Sub(status.StartedAt).Seconds())

	if err := b.saveSyncStatus(ctx, req.Storage, status); err != nil {
		return nil, err
	}

	if err := b.saveSyncHistory(ctx, req.Storage, status); err != nil {
		b.logger.Error("failed to save sync history", "error", err)
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"started_at":           status.StartedAt.Format(time.RFC3339),
			"status":               status.Status,
			"organizations_synced": status.OrganizationsSynced,
			"secrets_synced":       status.SecretsSynced,
			"failed":               status.Failed,
			"completed_at":         status.CompletedAt.Format(time.RFC3339),
			"duration_seconds":     status.DurationSeconds,
		},
	}, nil
}

func (b *Backend) pathSyncStatusRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	status, err := b.readSyncStatus(ctx, req.Storage)
	if err != nil {
		return nil, err
	}
	if status == nil {
		return &logical.Response{
			Data: map[string]interface{}{
				"last_sync":            nil,
				"last_org":             nil,
				"total_organizations":  0,
				"synced_organizations": 0,
				"total_secrets":        0,
				"synced_secrets":       0,
				"status":               "idle",
				"last_error":           nil,
			},
		}, nil
	}

	lastSync := ""
	if !status.LastSync.IsZero() {
		lastSync = status.LastSync.Format(time.RFC3339)
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"last_sync":            lastSync,
			"last_org":             status.LastOrg,
			"total_organizations":  status.TotalOrgs,
			"synced_organizations": status.SyncedOrgs,
			"total_secrets":        status.TotalSecrets,
			"synced_secrets":       status.SyncedSecrets,
			"status":               status.Status,
			"last_error":           status.LastError,
		},
	}, nil
}

func (b *Backend) pathSyncHistoryList(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	entry, err := req.Storage.Get(ctx, syncHistoryListKey)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return &logical.Response{
			Data: map[string]interface{}{
				"keys": []string{},
			},
		}, nil
	}

	var keys []string
	if err := entry.DecodeJSON(&keys); err != nil {
		return nil, err
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"keys": keys,
		},
	}, nil
}

func (b *Backend) pathSyncHistoryRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	entry, err := req.Storage.Get(ctx, syncHistoryListKey)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return logical.ErrorResponse("sync history entry not found"), logical.ErrInvalidRequest
	}

	var keys []string
	if err = entry.DecodeJSON(&keys); err != nil {
		return nil, err
	}

	if len(keys) == 0 {
		return &logical.Response{
			Data: map[string]interface{}{
				"keys": []string{},
			},
		}, nil
	}

	latestTimestamp := keys[len(keys)-1]
	path := syncHistoryPrefix + latestTimestamp
	historyEntry, err := req.Storage.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	if historyEntry == nil {
		return logical.ErrorResponse("sync history entry not found"), logical.ErrInvalidRequest
	}

	var history SyncHistoryEntry
	if err := historyEntry.DecodeJSON(&history); err != nil {
		return nil, err
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"timestamp":            history.Timestamp.Format(time.RFC3339),
			"status":               history.Status,
			"organizations_synced": history.OrganizationsSynced,
			"secrets_synced":       history.SecretsSynced,
			"failed":               history.Failed,
			"duration_seconds":     history.DurationSeconds,
		},
	}, nil
}

func (b *Backend) pathSyncHistoryTimestampRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	timestampRaw, ok := data.Get("timestamp").(string)
	if !ok {
		return logical.ErrorResponse("timestamp must be a string"), logical.ErrInvalidRequest
	}
	timestamp := timestampRaw
	if timestamp == "" {
		return logical.ErrorResponse("timestamp is required"), logical.ErrInvalidRequest
	}

	path := syncHistoryPrefix + timestamp
	entry, err := req.Storage.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return logical.ErrorResponse("sync history entry not found"), logical.ErrInvalidRequest
	}

	var history SyncHistoryEntry
	if err := entry.DecodeJSON(&history); err != nil {
		return nil, err
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"timestamp":            history.Timestamp.Format(time.RFC3339),
			"status":               history.Status,
			"organizations_synced": history.OrganizationsSynced,
			"secrets_synced":       history.SecretsSynced,
			"failed":               history.Failed,
			"duration_seconds":     history.DurationSeconds,
		},
	}, nil
}

func (b *Backend) saveSyncStatus(ctx context.Context, storage logical.Storage, status *SyncStatus) error {
	status.LastSync = time.Now().UTC()

	entry, err := logical.StorageEntryJSON(syncStatusPath, status)
	if err != nil {
		return err
	}

	return storage.Put(ctx, entry)
}

func (b *Backend) readSyncStatus(ctx context.Context, storage logical.Storage) (*SyncStatus, error) {
	entry, err := storage.Get(ctx, syncStatusPath)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, nil
	}

	var status SyncStatus
	if err := entry.DecodeJSON(&status); err != nil {
		return nil, err
	}

	return &status, nil
}

func (b *Backend) saveSyncHistory(ctx context.Context, storage logical.Storage, status *SyncStatus) error {
	entry := SyncHistoryEntry{
		Timestamp:           status.CompletedAt,
		Status:              status.Status,
		OrganizationsSynced: status.OrganizationsSynced,
		SecretsSynced:       status.SecretsSynced,
		Failed:              status.Failed,
		DurationSeconds:     status.DurationSeconds,
	}

	key := status.CompletedAt.Format(time.RFC3339)
	historyEntry, err := logical.StorageEntryJSON(syncHistoryPrefix+key, entry)
	if err != nil {
		return err
	}

	if err = storage.Put(ctx, historyEntry); err != nil {
		return err
	}

	keysEntry, err := storage.Get(ctx, syncHistoryListKey)
	var keys []string
	if err == nil && keysEntry != nil {
		if decodeErr := keysEntry.DecodeJSON(&keys); decodeErr != nil {
			return fmt.Errorf("failed to decode keys: %w", decodeErr)
		}
	}

	keys = append(keys, key)

	newKeysEntry, err := logical.StorageEntryJSON(syncHistoryListKey, keys)
	if err != nil {
		return err
	}

	return storage.Put(ctx, newKeysEntry)
}
