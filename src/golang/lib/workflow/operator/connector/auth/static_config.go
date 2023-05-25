package auth

import (
	"context"
	"encoding/json"
)

type StaticConfig struct {
	Conf map[string]string `json:"conf"`
}

// NewStaticConfig creates a new Config that directly uses the information
// in conf for authentication.
func NewStaticConfig(conf map[string]string) Config {
	return &StaticConfig{Conf: conf}
}

func (sc *StaticConfig) getType() configType {
	return staticConfigType
}

func (sc *StaticConfig) Marshal() ([]byte, error) {
	return json.Marshal(sc.Conf)
}

func (sc *StaticConfig) PublicConfig() map[string]string {
	publicConf := make(map[string]string, len(sc.Conf))

	// TODO: This is hacky for now. This is a union of sensitive fields
	// of configs over all integration types.
	sensitiveKeys := []string{
		"auth_uri",                    // MongoDB config.
		"password",                    // most integration configs have this field.
		"token",                       // slack, ECR config
		"service_account_credentials", // S3 config.
		"config_file_content",         // S3 config.
		"access_key_id",               // AWS config.
		"secret_access_key",           // AWS config.
		"expire_at",                   // ECR config.
		"service_account_key",         // GAR config.
	}

	for key, val := range sc.Conf {
		if !sliceContains(sensitiveKeys, key) {
			publicConf[key] = val
		}
	}

	return publicConf
}

func sliceContains(s []string, elem string) bool {
	for _, ss := range s {
		if ss == elem {
			return true
		}
	}
	return false
}

func (sc *StaticConfig) Refresh(ctx context.Context) (bool, error) {
	// staticConfig does not need to be refreshed
	return false, nil
}

// Set sets a key-value pair in the Config.
func (sc *StaticConfig) Set(key, value string) {
	sc.Conf[key] = value
}
