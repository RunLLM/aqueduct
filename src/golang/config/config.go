package config

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/dropbox/godropbox/errors"
	"gopkg.in/yaml.v2"
)

// globalConfigPath is set during Init
var globalConfigPath string

// globalConfig is initialized during Init and updated as changes are made to the fields
var globalConfig *ServerConfiguration

type ServerConfiguration struct {
	AqPath             string                `yaml:"aqPath"`
	EncryptionKey      string                `yaml:"encryptionKey"`
	RetentionJobPeriod string                `yaml:"retentionJobPeriod"`
	ApiKey             string                `yaml:"apiKey"`
	StorageConfig      *shared.StorageConfig `yaml:"storageConfig"`
}

// AqueductPath is the filepath to the Aqueduct installation.
func AqueductPath() string

// EncryptionKey is used for encrypting objects stored in the Aqueduct vault.
func EncryptionKey() string

// RetentionJobPeriod defines how long to wait before garbage collecting workflow runs.
func RetentionJobPeriod() string

// APIKey returns the API key the user must use when issueing requests.
func APIKey() string

// Storage returns the storage layer config.
func Storage() shared.StorageConfig

// UpdateStorage updates the storage layer config.
func UpdateStorage(newStorage *shared.StorageConfig) error

// Init initializes the global server configuration. It must be invoked before
// any config field is accessed, otherwise the value will be incorrect.
func Init(path string) error {
	globalConfigPath = path
	if err := loadConfig(); err != nil {
		return errors.Wrap(err, "Unable to initialize config. Please check that the config file is correctly formatted and retry.")
	}
}

// loadConfig reads the file at `configPath` into `globalConfig`.
func loadConfig() error {
	bts, err := ioutil.ReadFile(globalConfigPath)
	if err != nil {
		return err
	}

	var config ServerConfiguration
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
func dumpConfig()
