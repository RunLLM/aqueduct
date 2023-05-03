package shared

import (
	"database/sql/driver"

	"github.com/aqueducthq/aqueduct/lib/models/utils"
)

// IntegrationConfig contains credentials for a Resource
type IntegrationConfig map[string]string

func (c *IntegrationConfig) Value() (driver.Value, error) {
	return utils.ValueJSONB(*c)
}

func (c *IntegrationConfig) Scan(value interface{}) error {
	return utils.ScanJSONB(value, c)
}
