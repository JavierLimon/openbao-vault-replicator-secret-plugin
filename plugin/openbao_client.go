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

func (b *Backend) listSecretsInDestinationOrg(org string) ([]string, error) {
	ctx := context.Background()
	config, err := b.readConfig(ctx, b.storage)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}
	if config == nil {
		return nil, fmt.Errorf("configuration not found")
	}

	client := b.getOpenBaoClient(config.DestinationToken)
	if client == nil {
		return nil, fmt.Errorf("failed to create OpenBao client")
	}

	mount := config.DestinationMount
	if mount == "" {
		mount = "kv2"
	}

	org = strings.TrimSuffix(org, "/")
	org = strings.ReplaceAll(org, "//", "/")

	path := fmt.Sprintf("%s/metadata/%s/", mount, org)
	resp, err := client.Logical().List(path)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets at %s: %w", path, err)
	}
	if resp == nil {
		return []string{}, nil
	}

	keysRaw, ok := resp.Data["keys"]
	if !ok {
		return []string{}, nil
	}

	keys, ok := keysRaw.([]interface{})
	if !ok {
		return []string{}, nil
	}

	var secrets []string
	for _, k := range keys {
		if keyStr, ok := k.(string); ok {
			if strings.HasSuffix(keyStr, "/") {
				folderName := strings.TrimSuffix(keyStr, "/")
				subPath := org + "/" + folderName
				subSecrets, err := b.listSecretsRecursive(client, mount, subPath)
				if err != nil {
					return nil, err
				}
				for _, subSecret := range subSecrets {
					secrets = append(secrets, folderName+"/"+subSecret)
				}
			} else {
				secrets = append(secrets, keyStr)
			}
		}
	}

	return secrets, nil
}

func (b *Backend) listSecretsRecursive(client *api.Client, mount, currentPath string) ([]string, error) {
	path := fmt.Sprintf("%s/metadata/%s/", mount, currentPath)
	resp, err := client.Logical().List(path)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets at %s: %w", currentPath, err)
	}
	if resp == nil {
		return []string{}, nil
	}

	keysRaw, ok := resp.Data["keys"]
	if !ok {
		return []string{}, nil
	}

	keys, ok := keysRaw.([]interface{})
	if !ok {
		return []string{}, nil
	}

	var secrets []string
	for _, k := range keys {
		if keyStr, ok := k.(string); ok {
			if strings.HasSuffix(keyStr, "/") {
				folderName := strings.TrimSuffix(keyStr, "/")
				subPath := currentPath + "/" + folderName
				subSecrets, err := b.listSecretsRecursive(client, mount, subPath)
				if err != nil {
					return nil, err
				}
				for _, subSecret := range subSecrets {
					secrets = append(secrets, folderName+"/"+subSecret)
				}
			} else {
				secrets = append(secrets, keyStr)
			}
		}
	}

	return secrets, nil
}

func (b *Backend) deleteSecretFromDestination(org, secretName string) error {
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
	b.logger.Info("Deleting secret", "path", path)

	_, err = client.Logical().Delete(path)
	if err != nil {
		return fmt.Errorf("failed to delete secret at %s: %w", path, err)
	}

	b.logger.Info("Secret deleted successfully")
	return nil
}
