package engine

import (
	"context"
	"time"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/workflow/dag"
	"github.com/dropbox/godropbox/errors"
)

const (
	DefaultExecutionTimeout = 15 * time.Minute
	DefaultCleanupTimeout   = 2 * time.Minute
)

var (
	ErrOpExecSystemFailure       = errors.New("Operator execution failed due to system error.")
	ErrOpExecBlockingUserFailure = errors.New("Operator execution failed due to user error.")
)

type Engine interface {
	ScheduleWorkflow(ctx context.Context, workflowDag dag.WorkflowDag, workflowId string, name string, period string) error

	SyncWorkflow(ctx context.Context, workflowDag dag.WorkflowDag)

	ExecuteWorkflow(ctx context.Context, workflowDag dag.WorkflowDag) (shared.ExecutionStatus, error)

	// Cleanup is an end-of-orchestration hook meant to do any final cleanup work, after Execute completes.
	CleanupWorkflow(ctx context.Context, workflowDag dag.WorkflowDag)
}
