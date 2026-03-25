package replicator

import (
	"context"
	"fmt"
	"time"

	"github.com/openbao/openbao/sdk/v2/logical"
)

const (
	auditLogPrefix  = "audit/log/"
	auditLogListKey = "audit/log"
	auditBufferSize = 100
)

type AuditEventType string

const (
	AuditEventSyncStarted   AuditEventType = "sync_started"
	AuditEventSyncCompleted AuditEventType = "sync_completed"
	AuditEventSyncFailed    AuditEventType = "sync_failed"
)

type AuditEvent struct {
	Timestamp           time.Time      `json:"timestamp"`
	EventType           AuditEventType `json:"event_type"`
	Organization        string         `json:"organization,omitempty"`
	SecretPath          string         `json:"secret_path,omitempty"`
	Status              string         `json:"status"`
	Error               string         `json:"error,omitempty"`
	OrganizationsSynced int            `json:"organizations_synced,omitempty"`
	SecretsSynced       int            `json:"secrets_synced,omitempty"`
	DurationSeconds     int            `json:"duration_seconds,omitempty"`
	DryRun              bool           `json:"dry_run,omitempty"`
}

func (b *Backend) LogAuditEvent(ctx context.Context, storage logical.Storage, event *AuditEvent) error {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	key := fmt.Sprintf("%s%s", auditLogPrefix, event.Timestamp.Format(time.RFC3339Nano))

	entry, err := logical.StorageEntryJSON(key, event)
	if err != nil {
		return fmt.Errorf("failed to create audit entry: %w", err)
	}

	if err := storage.Put(ctx, entry); err != nil {
		return fmt.Errorf("failed to write audit entry: %w", err)
	}

	if err := b.appendAuditKey(ctx, storage, key); err != nil {
		b.logger.Warn("failed to append audit key to list", "error", err)
	}

	return nil
}

func (b *Backend) appendAuditKey(ctx context.Context, storage logical.Storage, key string) error {
	entry, err := storage.Get(ctx, auditLogListKey)
	var keys []string
	if err == nil && entry != nil {
		if decodeErr := entry.DecodeJSON(&keys); decodeErr != nil {
			return fmt.Errorf("failed to decode audit keys: %w", decodeErr)
		}
	}

	keys = append(keys, key)

	newEntry, err := logical.StorageEntryJSON(auditLogListKey, keys)
	if err != nil {
		return fmt.Errorf("failed to create audit list entry: %w", err)
	}

	return storage.Put(ctx, newEntry)
}

func (b *Backend) LogSyncStarted(ctx context.Context, storage logical.Storage, orgs []string, dryRun bool) error {
	event := &AuditEvent{
		EventType: AuditEventSyncStarted,
		Status:    "started",
		DryRun:    dryRun,
	}
	if len(orgs) > 0 {
		event.Organization = fmt.Sprintf("%d organizations requested", len(orgs))
	}
	return b.LogAuditEvent(ctx, storage, event)
}

func (b *Backend) LogSyncCompleted(ctx context.Context, storage logical.Storage, orgsSynced, secretsSynced, failedCount int, durationSeconds int) error {
	event := &AuditEvent{
		EventType:           AuditEventSyncCompleted,
		Status:              "completed",
		OrganizationsSynced: orgsSynced,
		SecretsSynced:       secretsSynced,
		DurationSeconds:     durationSeconds,
	}
	return b.LogAuditEvent(ctx, storage, event)
}

func (b *Backend) LogSyncFailed(ctx context.Context, storage logical.Storage, errMsg string) error {
	event := &AuditEvent{
		EventType: AuditEventSyncFailed,
		Status:    "failed",
		Error:     errMsg,
	}
	return b.LogAuditEvent(ctx, storage, event)
}

func (b *Backend) ListAuditLogs(ctx context.Context, storage logical.Storage) ([]string, error) {
	entry, err := storage.Get(ctx, auditLogListKey)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return []string{}, nil
	}

	var keys []string
	if err := entry.DecodeJSON(&keys); err != nil {
		return nil, err
	}

	return keys, nil
}

func (b *Backend) GetAuditLog(ctx context.Context, storage logical.Storage, timestamp string) (*AuditEvent, error) {
	key := auditLogPrefix + timestamp
	entry, err := storage.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, nil
	}

	var event AuditEvent
	if err := entry.DecodeJSON(&event); err != nil {
		return nil, err
	}

	return &event, nil
}
