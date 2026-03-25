package replicator

import (
	"context"
	"fmt"

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
	config, err := b.readConfigNoLock()
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

	path := fmt.Sprintf("%s/data/%s/%s", mount, org, secretName)
	_, err = client.Logical().Write(path, map[string]interface{}{
		"data": data,
	})
	if err != nil {
		return fmt.Errorf("failed to write secret at %s: %w", path, err)
	}

	return nil
}

func (b *Backend) readConfigNoLock() (*Configuration, error) {
	entry, err := b.storage.Get(context.Background(), configStoragePath)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, nil
	}

	var config Configuration
	if err := entry.DecodeJSON(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
