package dag

import (
	col_op "github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator"
)

func (dag *workflowDagImpl) ChecksWithError() []operator.Operator {
	operators := dag.Operators()
	results := make([]operator.Operator, 0, len(operators))
	for _, op := range operators {
		if op.Type() == col_op.CheckType && op.ExecState().HasBlockingFailure() {
			results = append(results, op)
		}
	}

	return results
}

func (dag *workflowDagImpl) ChecksWithWarning() []operator.Operator {
	operators := dag.Operators()
	results := make([]operator.Operator, 0, len(operators))
	for _, op := range operators {
		if op.Type() == col_op.CheckType && op.ExecState().HasWarning() {
			results = append(results, op)
		}
	}

	return results
}
