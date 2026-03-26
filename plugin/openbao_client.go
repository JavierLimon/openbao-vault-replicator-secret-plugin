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
	return b.writeToLocalKVWithMetadata(org, secretName, data, nil)
}

func (b *Backend) writeToLocalKVWithMetadata(org, secretName string, data map[string]interface{}, customMetadata map[string]interface{}) error {
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

	org = strings.TrimSuffix(org, "/")
	org = strings.ReplaceAll(org, "//", "/")

	path := fmt.Sprintf("%s/data/%s/%s", mount, org, secretName)
	b.logger.Info("Writing secret", "path", path)

	writeData := map[string]interface{}{
		"data": data,
	}

	if len(customMetadata) > 0 {
		writeData["custom_metadata"] = customMetadata
	}

	resp, err := client.Logical().Write(path, writeData)
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
