package replicator

import (
	"context"
	"fmt"
	"time"

	"github.com/openbao/openbao/sdk/v2/framework"
	"github.com/openbao/openbao/sdk/v2/logical"
)

const (
	auditLogPrefix          = "audit/log/"
	auditLogListKey         = "audit/logs"
	AuditEventSyncStarted   = "sync_started"
	AuditEventSyncCompleted = "sync_completed"
	AuditEventSyncFailed    = "sync_failed"
)

type AuditEvent struct {
	Timestamp     time.Time              `json:"timestamp"`
	OrgCount      int                    `json:"org_count"`
	SecretCount   int                    `json:"secret_count,omitempty"`
	DurationMs    int64                  `json:"duration_ms,omitempty"`
	EventType     string                 `json:"event_type"`
	RequestID     string                 `json:"request_id,omitempty"`
	Status        string                 `json:"status,omitempty"`
	Error         string                 `json:"error,omitempty"`
	Organizations []string               `json:"organizations,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

type AuditLogger struct {
	backend *Backend
}

func NewAuditLogger(b *Backend) *AuditLogger {
	return &AuditLogger{backend: b}
}

func (a *AuditLogger) LogSyncStarted(ctx context.Context, storage logical.Storage, orgs []string, requestID string) error {
	event := AuditEvent{
		Timestamp:     time.Now().UTC(),
		EventType:     AuditEventSyncStarted,
		RequestID:     requestID,
		Organizations: orgs,
		OrgCount:      len(orgs),
		Status:        "started",
	}

	if len(orgs) == 0 {
		event.Organizations = []string{"all"}
	}

	return a.writeAuditEvent(ctx, storage, event)
}

func (a *AuditLogger) LogSyncCompleted(ctx context.Context, storage logical.Storage, requestID string, orgCount, secretCount int, durationMs int64) error {
	event := AuditEvent{
		Timestamp:   time.Now().UTC(),
		EventType:   AuditEventSyncCompleted,
		RequestID:   requestID,
		OrgCount:    orgCount,
		SecretCount: secretCount,
		DurationMs:  durationMs,
		Status:      "completed",
	}

	return a.writeAuditEvent(ctx, storage, event)
}

func (a *AuditLogger) LogSyncFailed(ctx context.Context, storage logical.Storage, requestID string, err error) error {
	event := AuditEvent{
		Timestamp: time.Now().UTC(),
		EventType: AuditEventSyncFailed,
		RequestID: requestID,
		Status:    "failed",
	}

	if err != nil {
		event.Error = err.Error()
	}

	return a.writeAuditEvent(ctx, storage, event)
}

func (a *AuditLogger) writeAuditEvent(ctx context.Context, storage logical.Storage, event AuditEvent) error {
	key := fmt.Sprintf("%s%s", auditLogPrefix, event.Timestamp.Format(time.RFC3339Nano))

	entry, err := logical.StorageEntryJSON(key, event)
	if err != nil {
		return fmt.Errorf("failed to create storage entry: %w", err)
	}

	if err := storage.Put(ctx, entry); err != nil {
		return fmt.Errorf("failed to write audit event: %w", err)
	}

	if err := a.updateAuditLogIndex(ctx, storage, key); err != nil {
		return fmt.Errorf("failed to update audit log index: %w", err)
	}

	return nil
}

func (a *AuditLogger) updateAuditLogIndex(ctx context.Context, storage logical.Storage, key string) error {
	keysEntry, err := storage.Get(ctx, auditLogListKey)
	var keys []string
	if err == nil && keysEntry != nil {
		if decodeErr := keysEntry.DecodeJSON(&keys); decodeErr != nil {
			return fmt.Errorf("failed to decode keys: %w", decodeErr)
		}
	}

	keys = append(keys, key)

	newKeysEntry, err := logical.StorageEntryJSON(auditLogListKey, keys)
	if err != nil {
		return fmt.Errorf("failed to create keys entry: %w", err)
	}

	return storage.Put(ctx, newKeysEntry)
}

func (a *AuditLogger) ListAuditLogs(ctx context.Context, storage logical.Storage) ([]string, error) {
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

func (a *AuditLogger) GetAuditEvent(ctx context.Context, storage logical.Storage, timestamp string) (*AuditEvent, error) {
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

func (b *Backend) pathAuditLogs() *framework.Path {
	return &framework.Path{
		Pattern: "audit/logs/?",
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ListOperation: &framework.PathOperation{
				Summary:     "List audit logs",
				Description: "Lists all audit log entries",
				Callback:    b.pathAuditLogsList,
			},
			logical.ReadOperation: &framework.PathOperation{
				Summary:     "Get audit log entry",
				Description: "Retrieves a specific audit log entry by timestamp",
				Callback:    b.pathAuditLogRead,
			},
		},
		HelpSynopsis:    "List or get audit logs",
		HelpDescription: "Lists all audit log entries or retrieves a specific entry by timestamp",
	}
}

func (b *Backend) pathAuditLogsList(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	logger := NewAuditLogger(b)
	keys, err := logger.ListAuditLogs(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"keys": keys,
		},
	}, nil
}

func (b *Backend) pathAuditLogRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	timestampRaw, ok := data.Get("timestamp").(string)
	if !ok {
		return logical.ErrorResponse("timestamp is required"), logical.ErrInvalidRequest
	}

	logger := NewAuditLogger(b)
	event, err := logger.GetAuditEvent(ctx, req.Storage, timestampRaw)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return logical.ErrorResponse("audit log entry not found"), logical.ErrInvalidRequest
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"timestamp":    event.Timestamp.Format(time.RFC3339Nano),
			"event_type":   event.EventType,
			"request_id":   event.RequestID,
			"org_count":    event.OrgCount,
			"secret_count": event.SecretCount,
			"status":       event.Status,
			"error":        event.Error,
			"duration_ms":  event.DurationMs,
			"metadata":     event.Metadata,
		},
	}, nil
}
