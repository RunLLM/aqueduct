package workflow

import (
	"database/sql/driver"

	"github.com/aqueducthq/aqueduct/lib/collections/utils"
	"github.com/google/uuid"
)

type CronString string

type UpdateTrigger string

const (
	ManualUpdateTrigger    UpdateTrigger = "manual"
	PeriodicUpdateTrigger  UpdateTrigger = "periodic"
	AirflowUpdateTrigger   UpdateTrigger = "airflow"
	CascadingUpdateTrigger UpdateTrigger = "cascade"
)

type Schedule struct {
	Trigger              UpdateTrigger `json:"trigger"`
	CronSchedule         CronString    `json:"cron_schedule"`
	DisableManualTrigger bool          `json:"disable_manual_trigger"`
	Paused               bool          `json:"paused"`
	// SourceID is the source Workflow that triggers this
	// Workflow upon a successful run
	SourceID uuid.UUID `json:"source_id"`
}

func (s *Schedule) Value() (driver.Value, error) {
	return utils.ValueJsonB(*s)
}

func (s *Schedule) Scan(value interface{}) error {
	return utils.ScanJsonB(value, s)
}

type RetentionPolicy struct {
	KLatestRuns int `json:"k_latest_runs"`
}

func (s *RetentionPolicy) Value() (driver.Value, error) {
	return utils.ValueJsonB(*s)
}

func (s *RetentionPolicy) Scan(value interface{}) error {
	return utils.ScanJsonB(value, s)
}
