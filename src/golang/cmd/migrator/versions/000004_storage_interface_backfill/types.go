package _000004_storage_interface_backfill

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/aqueducthq/aqueduct/lib/models/utils"
)

type storageType string

const (
	s3StorageType storageType = "s3"
)

type storageConfig struct {
	Type       storageType `json:"type"`
	S3Config   *s3Config   `json:"s3_config,omitempty"`
	FileConfig *fileConfig `json:"file_config,omitempty"`
}

type fileConfig struct {
	Directory string `json:"directory"`
}

func (s *storageConfig) Scan(value interface{}) error {
	return utils.ScanJSONB(value, s)
}

func (s *storageConfig) Value() (driver.Value, error) {
	return utils.ValueJSONB(*s)
}

type s3Config struct {
	Region string `json:"region"`
	Bucket string `json:"bucket"`
}

func (s *s3Config) Value() (driver.Value, error) {
	return utils.ValueJSONB(*s)
}

func (s *s3Config) Scan(value interface{}) error {
	return utils.ScanJSONB(value, s)
}

type specType string

const (
	functionType   specType = "function"
	metricType     specType = "metric"
	validationType specType = "validation"
)

type specUnion struct {
	Type       specType    `json:"type"`
	Function   *function   `json:"function,omitempty"`
	Metric     *metric     `json:"metric,omitempty"`
	Validation *validation `json:"validation,omitempty"`
}

type spec struct {
	spec specUnion
}

func (s spec) getType() specType {
	return s.spec.Type
}

func (s spec) isFunction() bool {
	return s.getType() == functionType
}

func (s spec) Function() *function {
	if !s.isFunction() {
		return nil
	}

	return s.spec.Function
}

func (s spec) isMetric() bool {
	return s.getType() == metricType
}

func (s spec) Metric() *metric {
	if !s.isMetric() {
		return nil
	}

	return s.spec.Metric
}

func (s spec) isValidation() bool {
	return s.getType() == validationType
}

func (s spec) Validation() *validation {
	if !s.isValidation() {
		return nil
	}

	return s.spec.Validation
}

func (s spec) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.spec)
}

func (s *spec) UnmarshalJSON(rawMessage []byte) error {
	var spec specUnion
	err := json.Unmarshal(rawMessage, &spec)
	if err != nil {
		return err
	}

	// Overwrite the spec type based on the data.
	if spec.Function != nil {
		spec.Type = functionType
	} else if spec.Metric != nil {
		spec.Type = metricType
	} else if spec.Validation != nil {
		spec.Type = validationType
	}

	s.spec = spec
	return nil
}

func (s *spec) Value() (driver.Value, error) {
	return utils.ValueJSONB(*s)
}

func (s *spec) Scan(value interface{}) error {
	return utils.ScanJSONB(value, s)
}

type function struct {
	Type           fType           `json:"type"`
	Language       string          `json:"language"`
	Granularity    granularity     `json:"granularity"`
	S3Path         string          `json:"s3_path"`
	StoragePath    string          `json:"storage_path,omitempty"`
	GithubMetadata *githubMetadata `json:"github_metadata,omitempty"`
	EntryPoint     *entryPoint     `json:"entry_point,omitempty"`
	CustomArgs     string          `json:"custom_args"`
}

type fType string

type granularity string

type entryPoint struct {
	File      string `json:"file"`
	ClassName string `json:"class_name"`
	Method    string `json:"method"`
}

type metric struct {
	Function function `json:"function"`
}

type level string

type validation struct {
	Level    level    `json:"level"`
	Function function `json:"function"`
}

type githubRepoConfigContentType string

type githubMetadata struct {
	Owner  string `json:"owner"`
	Repo   string `json:"repo"`
	Branch string `json:"branch"`
	Path   string `json:"path"`

	RepoConfigContentType githubRepoConfigContentType `json:"repo_config_content_type,omitempty"`
	RepoConfigContentName string                      `json:"repo_config_content_name,omitempty"`

	RepoConfig *repoConfig `json:"repo_config,omitempty"`

	CommitId string `json:"commit_id"`
}

type operatorRepoConfig struct {
	Path       string `json:"path"`
	EntryPoint string `json:"entry_point"`
	ClassName  string `json:"class_name"`
	Method     string `json:"method"`
}

type queryRepoConfig struct {
	Path string `json:"path"`
}

type repoConfig struct {
	Operators map[string]operatorRepoConfig `json:"operators"`
	Queries   map[string]queryRepoConfig    `json:"queries"`
}
