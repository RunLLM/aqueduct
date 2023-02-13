package shared

import (
	"database/sql/driver"

	"github.com/aqueducthq/aqueduct/lib/models/utils"
	"github.com/google/uuid"
)

// CronString follows cron convention: https://en.wikipedia.org/wiki/Cron.
type CronString string

// UpdateTrigger specifies how a workflow can be invoked.
type UpdateTrigger string

const (
	ManualUpdateTrigger    UpdateTrigger = "manual"
	PeriodicUpdateTrigger  UpdateTrigger = "periodic"
	AirflowUpdateTrigger   UpdateTrigger = "airflow"
	CascadingUpdateTrigger UpdateTrigger = "cascade"
)

// Schedule defines the frequency for running a workflow.
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
	return utils.ValueJSONB(*s)
}

func (s *Schedule) Scan(value interface{}) error {
	return utils.ScanJSONB(value, s)
}
