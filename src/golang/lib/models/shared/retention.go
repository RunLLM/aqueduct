package shared

import (
	"database/sql/driver"

	"github.com/aqueducthq/aqueduct/lib/models/utils"
)

// A RetentionPolicy specifies that only KLatestRuns should be saved for each
// workflow.
type RetentionPolicy struct {
	KLatestRuns int `json:"k_latest_runs"`
}

func (r *RetentionPolicy) Value() (driver.Value, error) {
	return utils.ValueJSONB(*r)
}

func (r *RetentionPolicy) Scan(value interface{}) error {
	return utils.ScanJSONB(value, r)
}
