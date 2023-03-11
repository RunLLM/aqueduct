package auth

import (
	"context"
	"encoding/json"

	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/google/uuid"
)

const (
	secretConfigKey     = "config"
	secretConfigTypeKey = "config-type"
)

// WriteConfigToSecret takes a Config and stores it in a k8s secret.
// The name of the secret is integrationId.
func WriteConfigToSecret(
	ctx context.Context,
	integrationId uuid.UUID,
	config Config,
	vaultObject vault.Vault,
) error {
	// config is stored inside vault as a map[string]string as follows:
	// {
	//		secretConfigTypeKey: configType (e.g. staticConfigType, oauthConfigType)
	//		secretConfigKey: JSON encoded Config object (e.g. staticConfig, oauthConfig)
	// }
	secrets := make(map[string]string, 2)

	switch c := config.(type) {
	case *StaticConfig:
		secrets[secretConfigTypeKey] = string(staticConfigType)
	case *OAuthConfig:
		secrets[secretConfigTypeKey] = string(oauthConfigType)
	default:
		return errors.Newf("Unknown config type: %v", c)
	}

	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	secrets[secretConfigKey] = string(data)

	return vaultObject.Put(ctx, integrationId.String(), secrets)
}

// ReadConfigFromSecret reads a Config from the vault keyed by integrationId.
// It also refreshes the Config if necessary, and writes the updated Config to the k8s secret.
func ReadConfigFromSecret(
	ctx context.Context,
	integrationId uuid.UUID,
	vaultObject vault.Vault,
) (Config, error) {
	secrets, err := vaultObject.Get(ctx, integrationId.String())
	if err != nil {
		return nil, err
	}

	typStr, ok := secrets[secretConfigTypeKey]
	if !ok {
		// This means this secret has not yet been migrated to the new format,
		// so it must contain a staticConfig.
		return &StaticConfig{Conf: secrets}, nil
	}

	typ, err := parseConfigType(typStr)
	if err != nil {
		return nil, err
	}

	configData, ok := secrets[secretConfigKey]
	if !ok {
		return nil, errors.New("Malformed secret, missing config data.")
	}

	var config Config
	switch typ {
	case staticConfigType:
		config = &StaticConfig{}
	case oauthConfigType:
		config = &OAuthConfig{}
	default:
		return nil, errors.Newf("Unknown configType provided: %v", config)
	}

	if err := json.Unmarshal([]byte(configData), config); err != nil {
		return nil, err
	}

	// Refresh config if needed
	refresh, err := config.Refresh(ctx)
	if err != nil {
		return nil, err
	}

	if refresh {
		// The Config was refreshed, so the secret needs to be updated
		if err := WriteConfigToSecret(ctx, integrationId, config, vaultObject); err != nil {
			return nil, err
		}
	}

	return config, nil
}
