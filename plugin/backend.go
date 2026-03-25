package replicator

import (
	"context"
	"sync"

	"github.com/hashicorp/go-hclog"
	"github.com/openbao/openbao/sdk/v2/framework"
	"github.com/openbao/openbao/sdk/v2/logical"
)

type Backend struct {
	*framework.Backend
	storage logical.Storage
	mu      sync.RWMutex
	logger  hclog.Logger
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
		},
		BackendType: logical.TypeLogical,
	}

	if err := b.Setup(ctx, conf); err != nil {
		return nil, err
	}

	return b, nil
}

func configExistenceCheck(ctx context.Context, req *logical.Request, data *framework.FieldData) (bool, error) {
	storage := req.Storage
	entry, err := storage.Get(ctx, configStoragePath)
	if err != nil {
		return false, err
	}
	return entry != nil, nil
}
