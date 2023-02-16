package _000019_add_serialization_value_to_param_op

import (
	"database/sql/driver"

	"github.com/aqueducthq/aqueduct/lib/models/utils"
)

type Spec struct {
	Type  string            `db:"type" json:"type"`
	Param map[string]string `db:"param" json:"param"`
}

func (m *Spec) Value() (driver.Value, error) {
	return utils.ValueJSONB(*m)
}

func (m *Spec) Scan(value interface{}) error {
	return utils.ScanJSONB(value, m)
}
