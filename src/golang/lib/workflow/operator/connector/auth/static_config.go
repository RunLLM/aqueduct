package auth

import (
	"context"
	"encoding/json"
)

type staticConfig struct {
	Conf map[string]string `json:"conf"`
}

// NewStaticConfig creates a new Config that directly uses the information
// in conf for authentication.
func NewStaticConfig(conf map[string]string) Config {
	return &staticConfig{Conf: conf}
}

func (sc *staticConfig) getType() configType {
	return staticConfigType
}

func (sc *staticConfig) Marshal() ([]byte, error) {
	return json.Marshal(sc.Conf)
}

func (sc *staticConfig) PublicConfig() map[string]string {
	publicConf := make(map[string]string, len(sc.Conf))

	// TODO: This is hacky for now. It assumes the only confidential information
	// is "password" or "service_account_credentials" or "config_file_content".
	sensitiveKeys := []string{"password", "service_account_credentials", "config_file_content"}

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

func (sc *staticConfig) Refresh(ctx context.Context) (bool, error) {
	// staticConfig does not need to be refreshed
	return false, nil
}
