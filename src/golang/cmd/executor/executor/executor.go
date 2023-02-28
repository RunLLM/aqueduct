package executor

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/dropbox/godropbox/errors"
)

type Executor interface {
	// `Run` should execute the executor with the given context.
	Run(ctx context.Context) error
	// `Close` should terminate the execution and do all garbage collections.
	Close()
}

func NewExecutor(spec job.Spec) (Executor, error) {
	switch spec.Type() {
	case job.WorkflowJobType:
		workflowSpec, ok := spec.(*job.WorkflowSpec)
		if !ok {
			return nil, job.ErrInvalidJobSpec
		}

		base, err := NewBaseExecutor(workflowSpec.ExecutorConfig)
		if err != nil {
			return nil, err
		}
		return NewWorkflowExecutor(workflowSpec, base)
	case job.WorkflowRetentionType:
		workflowRetentionSpec, ok := spec.(*job.WorkflowRetentionSpec)
		if !ok {
			return nil, job.ErrInvalidJobSpec
		}
		base, err := NewBaseExecutor(workflowRetentionSpec.ExecutorConfig)
		if err != nil {
			return nil, err
		}

		return NewWorkflowRetentionExecutor(base), nil
	case job.DynamicTeardownType:
		dynamicTeardownSpec, ok := spec.(*job.DynamicTeardownSpec)
		if !ok {
			return nil, job.ErrInvalidJobSpec
		}
		base, err := NewBaseExecutor(dynamicTeardownSpec.ExecutorConfig)
		if err != nil {
			return nil, err
		}

		return NewDynamicTeardownExecutor(base), nil
	default:
		return nil, errors.New("Unsupported JobType")
	}
}
