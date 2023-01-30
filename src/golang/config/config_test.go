package config

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

const (
	testConfigPath = "test_config.yml"
)

var testConfig *serverConfiguration = &serverConfiguration{
	AqPath:             "/home/user/aqueduct",
	EncryptionKey:      "12345",
	RetentionJobPeriod: "* * * * *",
	ApiKey:             "user-api-key",
	StorageConfig: &shared.StorageConfig{
		Type:       shared.FileStorageType,
		FileConfig: &shared.FileConfig{},
	},
}

func TestInit(t *testing.T) {
	defer cleanup()
	setup(t)

	err := Init(testConfigPath)
	require.Nil(t, err)

	require.Equal(t, testConfigPath, globalConfigPath)
	require.True(t, reflect.DeepEqual(globalConfig, testConfig))
}

func TestUpdateStorage(t *testing.T) {
	defer cleanup()
	setup(t)

	err := Init(testConfigPath)
	require.Nil(t, err)

	currentStorage := Storage()
	require.True(t, reflect.DeepEqual(testConfig.StorageConfig, currentStorage))

	expectedStorage := &shared.StorageConfig{
		Type: shared.S3StorageType,
		S3Config: &shared.S3Config{
			Region:             "us-east-2",
			Bucket:             "test",
			CredentialsPath:    "/home/user/.aws",
			CredentialsProfile: "default",
		},
	}
	err = UpdateStorage(expectedStorage)
	require.Nil(t, err)

	actualStorage := Storage()
	require.True(t, reflect.DeepEqual(expectedStorage, actualStorage))
}

func TestLoadConfig(t *testing.T) {
	defer cleanup()
	setup(t)

	// Initialize globalConfigPath
	globalConfigPath = testConfigPath

	err := loadConfig()
	require.Nil(t, err)
}

func TestDumpConfig(t *testing.T) {
	defer cleanup()

	// Initialize globalConfig
	globalConfig = testConfig
	globalConfigPath = testConfigPath

	err := dumpConfig()
	require.Nil(t, err)
}

// setup creates a test config file
func setup(t *testing.T) {
	// Create a test config file
	data, err := yaml.Marshal(testConfig)
	require.Nil(t, err)

	err = ioutil.WriteFile(testConfigPath, data, 0o644)
	require.Nil(t, err)
}

// cleanup attempts to delete the test config file
func cleanup() {
	os.Remove(testConfigPath)
}
