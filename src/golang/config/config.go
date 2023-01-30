package config

import (
	"os"
	"path"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/dropbox/godropbox/errors"
	"gopkg.in/yaml.v2"
)

var (
	// globalConfigPath is set during Init
	globalConfigPath string
	// globalConfig is initialized during Init and updated as changes are made to the fields
	globalConfig *serverConfiguration
)

type serverConfiguration struct {
	AqPath             string                `yaml:"aqPath"`
	EncryptionKey      string                `yaml:"encryptionKey"`
	RetentionJobPeriod string                `yaml:"retentionJobPeriod"`
	ApiKey             string                `yaml:"apiKey"`
	StorageConfig      *shared.StorageConfig `yaml:"storageConfig"`
}

// AqueductPath is the filepath to the Aqueduct installation.
func AqueductPath() string {
	return globalConfig.AqPath
}

// EncryptionKey is used for encrypting objects stored in the Aqueduct vault.
func EncryptionKey() string {
	return globalConfig.EncryptionKey
}

// RetentionJobPeriod defines how long to wait before garbage collecting workflow runs.
func RetentionJobPeriod() string {
	return globalConfig.RetentionJobPeriod
}

// APIKey returns the API key the user must use when issuing requests.
func APIKey() string {
	return globalConfig.ApiKey
}

// Storage returns the storage layer config.
func Storage() *shared.StorageConfig {
	return globalConfig.StorageConfig
}

// UpdateStorage updates the storage layer config.
func UpdateStorage(newStorage *shared.StorageConfig) error {
	globalConfig.StorageConfig = newStorage
	return dumpConfig()
}

// Init initializes the global server configuration. It must be invoked before
// any config field is accessed, otherwise the value will be incorrect.
func Init(path string) error {
	globalConfigPath = path
	if err := loadConfig(); err != nil {
		return errors.Wrap(err, "Unable to initialize config. Please check that the config file is correctly formatted and retry.")
	}

	return nil
}

// loadConfig reads the file at `configPath` into `globalConfig`.
func loadConfig() error {
	bts, err := os.ReadFile(globalConfigPath)
	if err != nil {
		return err
	}

	var config serverConfiguration
	err = yaml.Unmarshal(bts, &config)
	if err != nil {
		return err
	}

	if config.StorageConfig == nil {
		// StorageConfig was not provided so the FileStorage is used by default
		defaultStoragePath := path.Join(os.Getenv("HOME"), ".aqueduct", "server", "storage")

		config.StorageConfig = &shared.StorageConfig{
			Type: shared.FileStorageType,
			FileConfig: &shared.FileConfig{
				Directory: defaultStoragePath,
			},
		}
	}

	globalConfig = &config

	return nil
}

// dumpConfig writes `globalConfig` to the file at `configPath`.
func dumpConfig() error {
	data, err := yaml.Marshal(globalConfig)
	if err != nil {
		return err
	}

	return os.WriteFile(globalConfigPath, data, 0o664)
}
