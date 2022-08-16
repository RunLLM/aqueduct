package integration

type S3ConfigType string

const (
	AccessKeyS3ConfigType         S3ConfigType = "access_key"
	ConfigFileS3ConfigType        S3ConfigType = "config_file"
	ConfigFileContentS3ConfigType S3ConfigType = "config_file_content"
)

// S3Config contains the fields for connecting an S3 integration.
type S3Config struct {
	Type              S3ConfigType `json:"type"`
	Bucket            string       `json:"bucket"`
	AccessKeyId       string       `json:"access_key_id"`
	SecretAccessKey   string       `json:"secret_access_key"`
	ConfigFilePath    string       `json:"config_file_path"`
	ConfigFileContent string       `json:"config_file_content"`
	ConfigFileProfile string       `json:"config_file_profile"`
}

// AirflowConfig contains the fields for connecting an Airflow integration.
type AirflowConfig struct{}
