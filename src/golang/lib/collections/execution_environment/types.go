package execution_environment

import (
	"database/sql/driver"

	"github.com/aqueducthq/aqueduct/lib/collections/utils"
)

type Spec struct {
	PythonVersion string   `json:"python_version,omitempty"`
	Dependencies  []string `json:"dependencies,omitempty"`
}

func (m *Spec) Value() (driver.Value, error) {
	return utils.ValueJsonB(*m)
}

func (m *Spec) Scan(value interface{}) error {
	return utils.ScanJsonB(value, m)
}
