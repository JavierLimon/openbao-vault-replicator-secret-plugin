package replicator

import (
	"context"

	"github.com/openbao/openbao/sdk/v2/framework"
	"github.com/openbao/openbao/sdk/v2/logical"
)

func (b *Backend) pathSync() *framework.Path {
	return &framework.Path{
		Pattern: "sync/secrets",
		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.CreateOperation: b.pathSyncSecrets,
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

func (b *Backend) pathSyncSecrets(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	return &logical.Response{
		Data: map[string]interface{}{
			"status": "not implemented yet",
		},
	}, nil
}
