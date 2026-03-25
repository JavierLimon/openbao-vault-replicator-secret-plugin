package replicator

import (
	"context"
	"sync"

	"github.com/hashicorp/go-hclog"
	"github.com/openbao/openbao/sdk/v2/framework"
	"github.com/openbao/openbao/sdk/v2/logical"
)

type Backend struct {
	storage logical.Storage
	logger  hclog.Logger
	*framework.Backend
	mu sync.RWMutex
}

func Factory(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {
	b := &Backend{
		storage: conf.StorageView,
		logger:  conf.Logger,
	}

	b.Backend = &framework.Backend{
		Help: "Vault Replicator - Secret engine plugin for replicating secrets from Vault to OpenBao",
		PathsSpecial: &logical.Paths{
			Unauthenticated: []string{"health"},
		},
		Paths: []*framework.Path{
			b.pathConfig(),
			b.pathSync(),
			b.pathSyncStatus(),
			b.pathSyncHistory(),
			b.pathSyncHistoryTimestamp(),
			b.pathAuditLogs(),
			b.pathHealth(),
		},
		BackendType: logical.TypeLogical,
	}

	if err := b.Setup(ctx, conf); err != nil {
		return nil, err
	}

	return b, nil
}
