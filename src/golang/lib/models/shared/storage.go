package shared

import (
	"database/sql/driver"

	"github.com/aqueducthq/aqueduct/lib/models/utils"
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

type S3Config struct {
	Region             string `yaml:"region" json:"region"`
	Bucket             string `yaml:"bucket" json:"bucket"`
	CredentialsPath    string `yaml:"credentialsPath" json:"credentials_path"`
	CredentialsProfile string `yaml:"credentialsProfile"  json:"credentials_profile"`
	AWSAccessKeyID     string `yaml:"awsAccessKeyId"  json:"aws_access_key_id"`
	AWSSecretAccessKey string `yaml:"awsSecretAccessKey"  json:"aws_secret_access_key"`
}

type FileConfig struct {
	Directory string `yaml:"directory" json:"directory"`
}

type GCSConfig struct {
	Bucket                    string `yaml:"bucket"  json:"bucket"`
	ServiceAccountCredentials string `yaml:"serviceAccountCredentials"  json:"service_account_credentials"`
}

func (s *StorageConfig) Scan(value interface{}) error {
	return utils.ScanJSONB(value, s)
}

func (s *StorageConfig) Value() (driver.Value, error) {
	return utils.ValueJSONB(*s)
}
