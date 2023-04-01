package shared

import (
	"database/sql/driver"

	"github.com/aqueducthq/aqueduct/lib/models/utils"
	"github.com/dropbox/godropbox/errors"
)

type StorageType string

const (
	S3StorageType   StorageType = "s3"
	FileStorageType StorageType = "file"
	GCSStorageType  StorageType = "gcs"
)

type StorageConfig struct {
	Type       StorageType `yaml:"type" json:"type"`
	S3Config   *S3Config   `yaml:"s3Config" json:"s3_config,omitempty"`
	FileConfig *FileConfig `yaml:"fileConfig" json:"file_config,omitempty"`
	GCSConfig  *GCSConfig  `yaml:"gcsConfig"  json:"gcs_config,omitempty"`
}

type StorageConfigPublic struct {
	Type            StorageType      `json:"type"`
	S3ConfigPublic  *S3ConfigPublic  `json:"s3Config,omitempty"`
	FileConfig      *FileConfig      `json:"fileConfig,omitempty"`
	GCSConfigPublic *GCSConfigPublic `json:"gcsConfig,omitempty"`

	// Empty means that the local filesystem is being used as storage.
	IntegrationName string `json:"integration_name,omitempty"`
}

type S3Config struct {
	Region             string `yaml:"region" json:"region"`
	Bucket             string `yaml:"bucket" json:"bucket"`
	CredentialsPath    string `yaml:"credentialsPath" json:"credentials_path"`
	CredentialsProfile string `yaml:"credentialsProfile"  json:"credentials_profile"`
	AWSAccessKeyID     string `yaml:"awsAccessKeyId"  json:"aws_access_key_id"`
	AWSSecretAccessKey string `yaml:"awsSecretAccessKey"  json:"aws_secret_access_key"`
}

type S3ConfigPublic struct {
	Region string `yaml:"region" json:"region"`
	Bucket string `yaml:"bucket" json:"bucket"`
}

type FileConfig struct {
	Directory string `yaml:"directory" json:"directory"`
}

type GCSConfig struct {
	Bucket                    string `yaml:"bucket"  json:"bucket"`
	ServiceAccountCredentials string `yaml:"serviceAccountCredentials"  json:"service_account_credentials"`
}

type GCSConfigPublic struct {
	Bucket string `yaml:"bucket"  json:"bucket"`
}

func (s *StorageConfig) Scan(value interface{}) error {
	return utils.ScanJSONB(value, s)
}

func (s *StorageConfig) Value() (driver.Value, error) {
	return utils.ValueJSONB(*s)
}

func (s *StorageConfig) ToPublic() (*StorageConfigPublic, error) {
	storageConfigPublic := &StorageConfigPublic{
		Type: s.Type,
	}

	switch s.Type {
	case FileStorageType:
		storageConfigPublic.FileConfig = s.FileConfig
	case S3StorageType:
		storageConfigPublic.S3ConfigPublic = &S3ConfigPublic{
			Region: s.S3Config.Region,
			Bucket: s.S3Config.Bucket,
		}
	case GCSStorageType:
		storageConfigPublic.GCSConfigPublic = &GCSConfigPublic{
			Bucket: s.GCSConfig.Bucket,
		}
	default:
		return nil, errors.Newf("Unknown storage type. %s", s.Type)
	}

	return storageConfigPublic, nil
}
