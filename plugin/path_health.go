package replicator

import (
	"context"

	"github.com/openbao/openbao/sdk/v2/framework"
	"github.com/openbao/openbao/sdk/v2/logical"
)

func (b *Backend) pathHealth() *framework.Path {
	return &framework.Path{
		Pattern: "health",
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ReadOperation: &framework.PathOperation{
				Summary:     "Health check",
				Description: "Returns the health status of the plugin",
				Callback:    b.pathHealthRead,
			},
		},
		HelpSynopsis:    "Health check endpoint",
		HelpDescription: "Returns the health status of the Vault Replicator plugin",
	}
}

func (b *Backend) pathHealthRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	return &logical.Response{
		Data: map[string]interface{}{
			"status": "ok",
			"plugin": "vault-replicator",
		},
	}, nil
}
