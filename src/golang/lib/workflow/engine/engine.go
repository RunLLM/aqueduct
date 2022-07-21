package engine

import (
	"context"
	"time"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
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
	Schedule(ctx context.Context, name string, period string)

	Sync(ctx context.Context)

	Execute(ctx context.Context) (shared.ExecutionStatus, error)

	// Finish is an end-of-orchestration hook meant to do any final cleanup work, after Execute completes.
	Finish(ctx context.Context)
}
