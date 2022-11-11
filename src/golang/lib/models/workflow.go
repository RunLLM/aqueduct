package models

import (
<<<<<<< HEAD
=======
	"fmt"
	"strings"
>>>>>>> 43305805ced645c2164adc119b46ef12a3843e17
	"time"

	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
)

<<<<<<< HEAD
// A Workflow maps to the workflow table.
type Workflow struct {
	ID              uuid.UUID              `db:"id" json:"id"`
	UserId          uuid.UUID              `db:"user_id" json:"user_id"`
=======
const (
	WorkflowTable = "workflow"

	// Workflow column names
	WorkflowID          = "id"
	WorkflowUserID      = "user_id"
	WorkflowName        = "name"
	WorkflowDescription = "description"
	WorkflowSchedule    = "schedule"
	WorkflowCreatedAt   = "created_at"
	WorkflowRetention   = "retention_policy"
)

// A Workflow maps to the workflow table.
type Workflow struct {
	ID              uuid.UUID              `db:"id" json:"id"`
	UserID          uuid.UUID              `db:"user_id" json:"user_id"`
>>>>>>> 43305805ced645c2164adc119b46ef12a3843e17
	Name            string                 `db:"name" json:"name"`
	Description     string                 `db:"description" json:"description"`
	Schedule        shared.Schedule        `db:"schedule" json:"schedule"`
	CreatedAt       time.Time              `db:"created_at" json:"created_at"`
	RetentionPolicy shared.RetentionPolicy `db:"retention_policy" json:"retention_policy"`
}
<<<<<<< HEAD
=======

// WorkflowCols returns a comma-separated string of all Workflow columns.
func WorkflowCols() string {
	return strings.Join(allWorkflowCols(), ",")
}

// WorkflowColsWithPrefix returns a comma-separated string of all
// Workflow columns prefixed by the table name.
func WorkflowColsWithPrefix() string {
	cols := allWorkflowCols()
	for i, col := range cols {
		cols[i] = fmt.Sprintf("%s.%s", WorkflowTable, col)
	}

	return strings.Join(cols, ",")
}

func allWorkflowCols() []string {
	return []string{
		WorkflowID,
		WorkflowUserID,
		WorkflowName,
		WorkflowDescription,
		WorkflowSchedule,
		WorkflowCreatedAt,
		WorkflowRetention,
	}
}
>>>>>>> 43305805ced645c2164adc119b46ef12a3843e17
