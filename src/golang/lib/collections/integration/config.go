package integration

import (
	"github.com/dropbox/godropbox/errors"
)

type S3ConfigType string

const (
	AccessKeyS3ConfigType         S3ConfigType = "access_key"
	ConfigFilePathS3ConfigType    S3ConfigType = "config_file_path"
	ConfigFileContentS3ConfigType S3ConfigType = "config_file_content"
)

type S3ConfigBool bool

func (scb *S3ConfigBool) UnmarshalJSON(data []byte) error {
	s := string(data)
	var b bool

	// TODO ENG-1586: Remove hack of treating credential string as a boolean
	switch s {
	case "\"true\"":
		b = true
	case "\"false\"":
		b = false
	default:
		return errors.Newf("Unable to unmarshal %s into S3ConfigBool", s)
	}

	*scb = S3ConfigBool(b)
	return nil
}

// S3Config contains the fields for connecting an S3 integration.
type S3Config struct {
	Type              S3ConfigType `json:"type"`
	Bucket            string       `json:"bucket"`
	Region            string       `json:"region"`
	AccessKeyId       string       `json:"access_key_id"`
	SecretAccessKey   string       `json:"secret_access_key"`
	ConfigFilePath    string       `json:"config_file_path"`
	ConfigFileContent string       `json:"config_file_content"`
	ConfigFileProfile string       `json:"config_file_profile"`
	UseAsStorage      S3ConfigBool `json:"use_as_storage"`
}

// AirflowConfig contains the fields for connecting an Airflow integration.
type AirflowConfig struct{}
