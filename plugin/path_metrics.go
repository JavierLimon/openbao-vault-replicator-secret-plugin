package replicator

import (
	"context"
	"sync/atomic"

	"github.com/openbao/openbao/sdk/v2/framework"
	"github.com/openbao/openbao/sdk/v2/logical"
)

type metrics struct {
	totalRequests     atomic.Int64
	totalErrors       atomic.Int64
	syncTotal         atomic.Int64
	syncCompleted     atomic.Int64
	syncFailed        atomic.Int64
	secretsReplicated atomic.Int64
}

func (b *Backend) pathMetrics() *framework.Path {
	return &framework.Path{
		Pattern: "metrics",
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ReadOperation: &framework.PathOperation{
				Summary:     "Metrics",
				Description: "Returns metrics for the plugin",
				Callback:    b.pathMetricsRead,
			},
		},
		HelpSynopsis:    "Metrics endpoint",
		HelpDescription: "Returns metrics for the Vault Replicator plugin",
	}
}

func (b *Backend) pathMetricsRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	return &logical.Response{
		Data: map[string]interface{}{
			"total_requests":     b.metrics.totalRequests.Load(),
			"total_errors":       b.metrics.totalErrors.Load(),
			"sync_total":         b.metrics.syncTotal.Load(),
			"sync_completed":     b.metrics.syncCompleted.Load(),
			"sync_failed":        b.metrics.syncFailed.Load(),
			"secrets_replicated": b.metrics.secretsReplicated.Load(),
		},
	}, nil
}

func (b *Backend) incrementTotalRequests() {
	b.metrics.totalRequests.Add(1)
}

func (b *Backend) incrementTotalErrors() {
	b.metrics.totalErrors.Add(1)
}

func (b *Backend) incrementSyncTotal() {
	b.metrics.syncTotal.Add(1)
}

func (b *Backend) incrementSyncCompleted() {
	b.metrics.syncCompleted.Add(1)
}

func (b *Backend) incrementSyncFailed() {
	b.metrics.syncFailed.Add(1)
}

func (b *Backend) incrementSecretsReplicated(count int64) {
	b.metrics.secretsReplicated.Add(count)
}
