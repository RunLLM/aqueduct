package _000017_update_to_canceled_status

import (
	"context"
	"database/sql/driver"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models/utils"
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

type ExecutionState struct {
	UserLogs *Logs           `json:"user_logs"`
	Status   ExecutionStatus `json:"status"`

	// These fields are only set if status == Failed.
	FailureType *FailureType `json:"failure_type"`
	Error       *Error       `json:"error"`
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

type artifactStatusInformation struct {
	ArtifactResultID uuid.UUID          `db:"id"`
	ExecState        NullExecutionState `db:"execution_state"`
}

type operatorStatusInformation struct {
	OperatorResultID uuid.UUID          `db:"id"`
	ExecState        NullExecutionState `db:"execution_state"`
}

func getPendingArtifactResultStatuses(ctx context.Context, db database.Database) ([]artifactStatusInformation, error) {
	// We're guaranteed that all pending artifacts are failed because when the
	// migration runs, the server is not running, so there will be no in-flight
	// runs.
	query := fmt.Sprintf("SELECT id, execution_state FROM artifact_result WHERE status=\"%s\";", PendingExecutionStatus)

	var artifactIds []artifactStatusInformation
	err := db.Query(ctx, &artifactIds, query)

	return artifactIds, err
}

func getFailedArtifactResultStatuses(ctx context.Context, db database.Database) ([]artifactStatusInformation, error) {
	query := fmt.Sprintf("SELECT id, execution_state FROM artifact_result WHERE status=\"%s\";", FailedExecutionStatus)

	var artifactIds []artifactStatusInformation
	err := db.Query(ctx, &artifactIds, query)

	return artifactIds, err
}

func getPendingOperatorResultStatuses(ctx context.Context, db database.Database) ([]operatorStatusInformation, error) {
	// We're guaranteed that all pending operators are failed because when the
	// migration runs, the server is not running, so there will be no in-flight
	// runs.
	query := fmt.Sprintf("SELECT id, execution_state FROM operator_result WHERE status=\"%s\";", PendingExecutionStatus)

	var operatorIds []operatorStatusInformation
	err := db.Query(ctx, &operatorIds, query)

	return operatorIds, err
}

func getCanceledArtifactResultStatuses(ctx context.Context, db database.Database) ([]artifactStatusInformation, error) {
	query := fmt.Sprintf("SELECT id, execution_state FROM artifact_result WHERE status=\"%s\";", CanceledExecutionStatus)

	var artifactIds []artifactStatusInformation
	err := db.Query(ctx, &artifactIds, query)

	return artifactIds, err
}

func getCanceledOperatorResultStatuses(ctx context.Context, db database.Database) ([]operatorStatusInformation, error) {
	query := fmt.Sprintf("SELECT id, execution_state FROM operator_result WHERE status=\"%s\";", CanceledExecutionStatus)

	var operatorIds []operatorStatusInformation
	err := db.Query(ctx, &operatorIds, query)

	return operatorIds, err
}
