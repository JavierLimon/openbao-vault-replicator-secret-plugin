package replicator

import (
	"context"

	"github.com/openbao/openbao/sdk/v2/framework"
	"github.com/openbao/openbao/sdk/v2/logical"
)

func (b *Backend) pathRoles() *framework.Path {
	return &framework.Path{
		Pattern: "roles",
		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.ReadOperation: b.pathRolesRead,
			logical.ListOperation: b.pathRolesList,
		},
		HelpSynopsis:    "Roles endpoint for Vault Replicator plugin",
		HelpDescription: "Manage roles for secret replication",
	}
}

func (b *Backend) pathRolesRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	return nil, logical.ErrUnsupportedOperation
}

func (b *Backend) pathRolesList(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	return nil, logical.ErrUnsupportedOperation
}
