package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aqueducthq/aqueduct/config"
	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
)

func ErrObjectDoesNotExist() error {
	return errors.New("Object does not exist in storage.")
}

type Storage interface {
	// Throws `ErrObjectDoesNotExist` if the path does not exist.
	Get(ctx context.Context, key string) ([]byte, error)
	Put(ctx context.Context, key string, value []byte) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) bool
}

func NewStorage(config *shared.StorageConfig) Storage {
	if config == nil {
		log.Fatalf("Nil storage config.")
	}

	switch config.Type {
	case shared.S3StorageType:
		return newS3Storage(config.S3Config)
	case shared.FileStorageType:
		return newFileStorage(config.FileConfig)
	case shared.GCSStorageType:
		return newGCSStorage(config.GCSConfig)
	default:
		log.Fatalf("Unsupported storage type: %s", config.Type)
		return nil
	}
}

func convertS3IntegrationtoStorageConfig(c *shared.S3IntegrationConfig) (*shared.StorageConfig, error) {
	// Users provide AWS credentials for an S3 integration via one of the following:
	//  1. AWS Access Key and Secret Key
	//  2. Credentials file content
	//  3. Credentials filepath and profile name
	// The S3 Storage implementation expects the AWS credentials to be specified via a
	// filepath and profile name, so we must convert the above to the correct format.
	storageConfig := &shared.StorageConfig{
		Type: shared.S3StorageType,
		S3Config: &shared.S3Config{
			Bucket: fmt.Sprintf("s3://%s", c.Bucket),
			Region: c.Region,
		},
	}
	switch c.Type {
	case shared.AccessKeyS3ConfigType:
		// AWS access and secret keys need to be written to a credentials file
		path := filepath.Join(config.AqueductPath(), "storage", uuid.NewString())
		f, err := os.Create(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		credentialsContent := fmt.Sprintf(
			"[default]\naws_access_key_id=%s\naws_secret_access_key=%s\n",
			c.AccessKeyId,
			c.SecretAccessKey,
		)
		if _, err := f.WriteString(credentialsContent); err != nil {
			return nil, err
		}

		storageConfig.S3Config.CredentialsPath = path
		storageConfig.S3Config.CredentialsProfile = "default"
	case shared.ConfigFileContentS3ConfigType:
		// The credentials content needs to be written to a credentials file
		path := filepath.Join(config.AqueductPath(), "storage", uuid.NewString())
		f, err := os.Create(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		// Determine profile name by looking for [profile_name]
		i := strings.Index(c.ConfigFileContent, "[")
		if i < 0 {
			return nil, errors.New("Unable to determine AWS credentials profile name.")
		}

		j := strings.Index(c.ConfigFileContent, "]")
		if j < 0 {
			return nil, errors.New("Unable to determine AWS credentials profile name.")
		}

		profileName := c.ConfigFileContent[i+1 : j]

		if _, err := f.WriteString(c.ConfigFileContent); err != nil {
			return nil, err
		}

		storageConfig.S3Config.CredentialsPath = path
		storageConfig.S3Config.CredentialsProfile = profileName
	case shared.ConfigFilePathS3ConfigType:
		// The credentials are already in the form of a filepath and profile, so no changes
		// need to be made
		storageConfig.S3Config.CredentialsPath = c.ConfigFilePath
		storageConfig.S3Config.CredentialsProfile = c.ConfigFileProfile
	default:
		return nil, errors.Newf("Unknown S3ConfigType: %v", c.Type)
	}

	return storageConfig, nil
}

func convertGCSIntegrationtoStorageConfig(c *shared.GCSIntegrationConfig) *shared.StorageConfig {
	return &shared.StorageConfig{
		Type: shared.GCSStorageType,
		GCSConfig: &shared.GCSConfig{
			Bucket:                    c.Bucket,
			ServiceAccountCredentials: c.ServiceAccountCredentials,
		},
	}
}

func ConvertIntegrationConfigToStorageConfig(
	svc shared.Service,
	confData []byte,
) (*shared.StorageConfig, error) {
	switch svc {
	case shared.S3:
		var c shared.S3IntegrationConfig
		if err := json.Unmarshal(confData, &c); err != nil {
			return nil, err
		}

		return convertS3IntegrationtoStorageConfig(&c)
	case shared.GCS:
		var c shared.GCSIntegrationConfig
		if err := json.Unmarshal(confData, &c); err != nil {
			return nil, err
		}

		return convertGCSIntegrationtoStorageConfig(&c), nil
	default:
		return nil, errors.Newf("%v cannot be used as the storage layer", svc)
	}
}
