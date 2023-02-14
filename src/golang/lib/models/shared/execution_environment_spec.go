package shared

import (
	"database/sql/driver"

	"github.com/aqueducthq/aqueduct/lib/models/utils"
)

type ExecutionEnvironmentSpec struct {
	PythonVersion string   `json:"python_version,omitempty"`
	Dependencies  []string `json:"dependencies,omitempty"`
}

func (m *ExecutionEnvironmentSpec) Value() (driver.Value, error) {
	return utils.ValueJSONB(*m)
}

func (m *ExecutionEnvironmentSpec) Scan(value interface{}) error {
	return utils.ScanJSONB(value, m)
}
