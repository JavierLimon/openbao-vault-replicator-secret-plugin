package replicator

import (
	"context"
	"sync"

	"github.com/hashicorp/go-hclog"
	"github.com/openbao/openbao/sdk/v2/framework"
	"github.com/openbao/openbao/sdk/v2/logical"
)

// Backend implements plugin.Interface
type Backend struct {
	*framework.Backend
	storage logical.Storage
	
	// Thread safety
	mu sync.RWMutex
	logger hclog.Logger
}

// Factory creates a new backend
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
		},
		BackendType: logical.TypeLogical,
	}
	
	if err := b.Setup(ctx, conf); err != nil {
		return nil, err
	}
	
	return b, nil
}
