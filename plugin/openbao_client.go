package replicator

import (
	"context"
	"fmt"
	"strings"

	"github.com/openbao/openbao/api/v2"
)

func (b *Backend) getOpenBaoClient(token string) *api.Client {
	config := api.DefaultConfig()

	client, err := api.NewClient(config)
	if err != nil {
		b.logger.Error("failed to create OpenBao client", "error", err)
		return nil
	}
	client.SetToken(token)
	return client
}

func (b *Backend) writeToLocalKV(org, secretName string, data map[string]interface{}) error {
	ctx := context.Background()
	config, err := b.readConfig(ctx, b.storage)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}
	if config == nil {
		return fmt.Errorf("configuration not found")
	}

	client := b.getOpenBaoClient(config.DestinationToken)
	if client == nil {
		return fmt.Errorf("failed to create OpenBao client")
	}

	mount := config.DestinationMount
	if mount == "" {
		mount = "kv2"
	}

	// Trim trailing slash from org to prevent double slashes
	org = strings.TrimSuffix(org, "/")
	// Also clean up any double slashes in the path
	org = strings.ReplaceAll(org, "//", "/")

	path := fmt.Sprintf("%s/data/%s/%s", mount, org, secretName)
	b.logger.Info("Writing secret", "path", path)

	resp, err := client.Logical().Write(path, map[string]interface{}{
		"data": data,
	})
	if err != nil {
		return fmt.Errorf("failed to write secret at %s: %w", path, err)
	}

	b.logger.Info("Secret written successfully", "response", resp)

	return nil
}

func (b *Backend) readConfigNoLock() (*Configuration, error) {
	ctx := context.Background()
	return b.readConfig(ctx, b.storage)
}
