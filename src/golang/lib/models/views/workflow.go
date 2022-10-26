package views

import (
	"time"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/collections/utils"
	"github.com/google/uuid"
)

// A LatestWorkflowStatus defines the status of the last run of a Workflow
// and additional Workflow metadata.
type LatestWorkflowStatus struct {
	ID          uuid.UUID                  `db:"id" json:"id"`
	Name        string                     `db:"name" json:"name"`
	Description string                     `db:"description" json:"description"`
	CreatedAt   time.Time                  `db:"created_at" json:"created_at"`
	LastRunAt   utils.NullTime             `db:"last_run_at" json:"last_run_at"`
	Status      shared.NullExecutionStatus `db:"status" json:"status"`
	Engine      string                     `db:"engine" json:"engine"`
}
