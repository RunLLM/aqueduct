package _000018_add_dag_result_exec_state_column

import (
	"context"
	"database/sql/driver"
	"time"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models/utils"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/google/uuid"
)

type ExecutionStatus string

const (
	SucceededExecutionStatus ExecutionStatus = "succeeded"
	FailedExecutionStatus    ExecutionStatus = "failed"
	RunningExecutionStatus   ExecutionStatus = "running"
	PendingExecutionStatus   ExecutionStatus = "pending"
	UnknownExecutionStatus   ExecutionStatus = "unknown"
	CanceledExecutionStatus  ExecutionStatus = "canceled"
)

type FailureType int64

const (
	Success          FailureType = 0
	SystemFailure    FailureType = 1
	UserFatalFailure FailureType = 2

	// Orchestration can continue onwards, despite this failure.
	// Eg. Check operator with WARNING severity does not pass.
	UserNonFatalFailure FailureType = 3
)

type Logs struct {
	Stdout string `json:"stdout"`
	StdErr string `json:"stderr"`
}

type Error struct {
	Context string `json:"context"`
	Tip     string `json:"tip"`
}

type ExecutionTimestamps struct {
	RegisteredAt *time.Time `json:"registered_at"`
	PendingAt    *time.Time `json:"pending_at"`
	RunningAt    *time.Time `json:"running_at"`
	FinishedAt   *time.Time `json:"finished_at"`
}

type ExecutionState struct {
	UserLogs *Logs           `json:"user_logs"`
	Status   ExecutionStatus `json:"status"`

	// These fields are only set if status == Failed.
	FailureType *FailureType         `json:"failure_type"`
	Error       *Error               `json:"error"`
	Timestamps  *ExecutionTimestamps `json:"timestamps"`
}

func (e *ExecutionState) Value() (driver.Value, error) {
	return utils.ValueJSONB(*e)
}

func (e *ExecutionState) Scan(value interface{}) error {
	return utils.ScanJSONB(value, e)
}

type NullExecutionState struct {
	ExecutionState
	IsNull bool
}

func (n *NullExecutionState) Value() (driver.Value, error) {
	if n.IsNull {
		return nil, nil
	}

	return (&n.ExecutionState).Value()
}

func (n *NullExecutionState) Scan(value interface{}) error {
	if value == nil {
		n.IsNull = true
		return nil
	}

	logs := &ExecutionState{}
	if err := logs.Scan(value); err != nil {
		return err
	}

	n.ExecutionState, n.IsNull = *logs, false
	return nil
}

type partialWorkflowDagResult struct {
	Id        uuid.UUID       `db:"id"`
	Status    ExecutionStatus `db:"status"`
	CreatedAt time.Time       `db:"created_at"`
}

func getAllDagResults(ctx context.Context, db database.Database) ([]partialWorkflowDagResult, error) {
	query := "SELECT id, status, created_at FROM workflow_dag_result;"

	var results []partialWorkflowDagResult
	err := db.Query(ctx, &results, query)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func backfill(ctx context.Context, db database.Database) error {
	results, err := getAllDagResults(ctx, db)
	if err != nil {
		return err
	}

	for _, dagResult := range results {
		changes := map[string]interface{}{
			"execution_state": &ExecutionState{
				Status: dagResult.Status,
				Timestamps: &ExecutionTimestamps{
					PendingAt: &dagResult.CreatedAt,
				},
			},
		}

		err = repos.UpdateRecord(ctx, changes, "workflow_dag_result", "id", dagResult.Id, db)
		if err != nil {
			return err
		}
	}

	return nil
}
