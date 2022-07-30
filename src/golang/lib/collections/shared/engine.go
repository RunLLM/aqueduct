package shared

import (
	"database/sql/driver"

	"github.com/aqueducthq/aqueduct/lib/collections/utils"
)

type EngineType string

const (
	AqueductEngineType EngineType = "aqueduct"
)

type EngineConfig struct {
	Type           EngineType      `yaml:"type" json:"type"`
	AqueductConfig *AqueductConfig `yaml:"aqueductConfig" json:"aqueduct_config,omitempty"`
}

type AqueductConfig struct{}

func (e *EngineConfig) Scan(value interface{}) error {
	return utils.ScanJsonB(value, e)
}

func (e *EngineConfig) Value() (driver.Value, error) {
	return utils.ValueJsonB(*e)
}
