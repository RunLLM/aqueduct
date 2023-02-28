package engine

import (
	"context"
	"time"

	"github.com/aqueducthq/aqueduct/lib/database"
	exec_env "github.com/aqueducthq/aqueduct/lib/execution_environment"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/models/shared/operator/param"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

const (
	DefaultExecutionTimeout     = 20 * time.Minute
	DefaultCleanupTimeout       = 2 * time.Minute
	DefaultPollIntervalMillisec = 300
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
		parameters map[string]param.Param,
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
		schedule *shared.Schedule,
		retentionPolicy *shared.RetentionPolicy,
		notificationSettings *shared.NotificationSettings,
	) error

	// TODO ENG-1444: Used as a wrapper to trigger a workflow via executor binary.
	// Remove once executor is removed.
	TriggerWorkflow(
		ctx context.Context,
		workflowId uuid.UUID,
		name string,
		timeConfig *AqueductTimeConfig,
		parameters map[string]param.Param,
	) (shared.ExecutionStatus, error)

	// InitEnv initialize the given environment for this engine.
	// This typically involves time-consuming steps that we want to avoid
	// during execution time, like creating conda or docker img.
	InitEnv(ctx context.Context, env *exec_env.ExecutionEnvironment) error
}

// AqEngine should be implemented by aqEngine
// which is used by all aqueduct-orchestrated engines.
type AqEngine interface {
	Engine

	PreviewWorkflow(
		ctx context.Context,
		dbDAG *models.DAG,
		execEnvByOperatorId map[uuid.UUID]exec_env.ExecutionEnvironment,
		timeConfig *AqueductTimeConfig,
	) (*WorkflowPreviewResult, error)
}

// SelfOrchestratedEngine should be implemented for each self-orchestrated engine.
// ie airflowEngine, rayEngine
type SelfOrchestratedEngine interface {
	Engine

	SyncWorkflow(
		ctx context.Context,
		dag *models.DAG,
	)
}
