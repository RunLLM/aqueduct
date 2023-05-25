package shared

import (
	"database/sql/driver"

	"github.com/aqueducthq/aqueduct/lib/models/utils"
)

// ResourceConfig contains credentials for a Resource
type ResourceConfig map[string]string

func (c *ResourceConfig) Value() (driver.Value, error) {
	return utils.ValueJSONB(*c)
}

func (c *ResourceConfig) Scan(value interface{}) error {
	return utils.ScanJSONB(value, c)
}
