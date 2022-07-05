package shared

import (
	"database/sql/driver"

	"github.com/aqueducthq/aqueduct/lib/collections/utils"
)

type StorageType string

const (
	S3StorageType   StorageType = "s3"
	FileStorageType StorageType = "file"
)

type StorageConfig struct {
	Type       StorageType `yaml:"type" json:"type"`
	S3Config   *S3Config   `yaml:"s3Config" json:"s3_config,omitempty"`
	FileConfig *FileConfig `yaml:"fileConfig" json:"file_config,omitempty"`
}

type S3Config struct {
	Region             string `yaml:"region" json:"region"`
	Bucket             string `yaml:"bucket" json:"bucket"`
	CredentialsPath    string `yaml:"credentialsPath" json:"credentials_path"`
	CredentialsProfile string `json:"credentialsProfile"  yaml:"credentials_profile"`
}

type FileConfig struct {
	Directory string `yaml:"directory" json:"directory"`
}

func (s *StorageConfig) Scan(value interface{}) error {
	return utils.ScanJsonB(value, s)
}

func (s *StorageConfig) Value() (driver.Value, error) {
	return utils.ValueJsonB(*s)
}
