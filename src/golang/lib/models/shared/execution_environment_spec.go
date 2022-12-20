package shared

import (
	"database/sql/driver"

	"github.com/aqueducthq/aqueduct/lib/collections/utils"
)

type ExecutionEnvironmentSpec struct {
	PythonVersion string   `json:"python_version,omitempty"`
	Dependencies  []string `json:"dependencies,omitempty"`
}

func (m *ExecutionEnvironmentSpec) Value() (driver.Value, error) {
	return utils.ValueJsonB(*m)
}

func (m *ExecutionEnvironmentSpec) Scan(value interface{}) error {
	return utils.ScanJsonB(value, m)
}
