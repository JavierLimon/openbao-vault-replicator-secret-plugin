package replicator

import (
	"context"

	"github.com/openbao/openbao/sdk/v2/framework"
	"github.com/openbao/openbao/sdk/v2/logical"
)

func (b *Backend) pathSync() *framework.Path {
	return &framework.Path{
		Pattern: "sync",
		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.ReadOperation:   b.pathSyncRead,
			logical.CreateOperation: b.pathSyncCreate,
		},
		HelpSynopsis:    "Sync endpoint for Vault Replicator plugin",
		HelpDescription: "Trigger and monitor secret replication from Vault to OpenBao",
	}
}

func (b *Backend) pathSyncRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	return nil, logical.ErrUnsupportedOperation
}

func (b *Backend) pathSyncCreate(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	return nil, logical.ErrUnsupportedOperation
}
