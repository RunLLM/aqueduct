package engine

import (
	"context"
	"time"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

const (
	DefaultExecutionTimeout     = 15 * time.Minute
	DefaultCleanupTimeout       = 2 * time.Minute
	DefaultPollIntervalMillisec = 100
)

var (
	ErrOpExecSystemFailure       = errors.New("Operator execution failed due to system error.")
	ErrOpExecBlockingUserFailure = errors.New("Operator execution failed due to user error.")
)

type Engine interface {
	ScheduleWorkflow(
		ctx context.Context,
		workflowId uuid.UUID,
		name string,
		period string,
	) error

	ExecuteWorkflow(
		ctx context.Context,
		workflowId uuid.UUID,
		timeConfig *AqueductTimeConfig,
		parameters map[string]string,
	) (shared.ExecutionStatus, error)

	DeleteWorkflow(
		ctx context.Context,
		workflowId uuid.UUID,
	) error

	EditWorkflow(
		ctx context.Context,
		txn database.Database,
		workflowId uuid.UUID,
		workflowName string,
		workflowDescription string,
		schedule *workflow.Schedule,
	) error
}

// AqEngine should be implemented by aqEngine
// which is used by all aqueduct-orchestrated engines.
type AqEngine interface {
	Engine

	PreviewWorkflow(
		ctx context.Context,
		dbWorkflowDag *workflow_dag.DBWorkflowDag,
		timeConfig *AqueductTimeConfig,
	) (*WorkflowPreviewResult, error)
}

// SelfOrchestratedEngine should be implemented for each self-orchestrated engine.
// ie airflowEngine, rayEngine
type SelfOrchestratedEngine interface {
	Engine

	SyncWorkflow(
		ctx context.Context,
		dbWorkflowDag *workflow_dag.DBWorkflowDag,
	)
}
