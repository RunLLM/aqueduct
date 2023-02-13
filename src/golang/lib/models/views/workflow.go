package views

import (
	"time"

	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/models/utils"
	"github.com/google/uuid"
)

// LatestWorkflowStatus defines the status of the last run of a Workflow
// and additional Workflow metadata.
type LatestWorkflowStatus struct {
	ID          uuid.UUID                  `db:"id" json:"id"`
	DagID       uuid.UUID                  `db:"dag_id" json:"dag_id"`
	ResultID    uuid.UUID                  `db:"result_id" json:"result_id"`
	Name        string                     `db:"name" json:"name"`
	Description string                     `db:"description" json:"description"`
	CreatedAt   time.Time                  `db:"created_at" json:"created_at"`
	LastRunAt   utils.NullTime             `db:"last_run_at" json:"last_run_at"`
	Status      shared.NullExecutionStatus `db:"status" json:"status"`
	Engine      string                     `db:"engine" json:"engine"`
}

// WorkflowLastRun is a wrapper around the last run at time for a Workflow
// and additional metadata.
type WorkflowLastRun struct {
	ID        uuid.UUID       `db:"workflow_id" json:"workflow_id"`
	Schedule  shared.Schedule `db:"schedule" json:"schedule"`
	LastRunAt time.Time       `db:"last_run_at" json:"last_run_at"`
}
