package dag

import (
	"github.com/aqueducthq/aqueduct/lib/workflow/operator"
)

func (dag *workflowDagImpl) OperatorsWithError() []operator.Operator {
	operators := dag.Operators()
	results := make([]operator.Operator, 0, len(operators))
	for _, op := range operators {
		if op.ExecState().HasBlockingFailure() {
			results = append(results, op)
		}
	}

	return results
}

func (dag *workflowDagImpl) OperatorsWithWarning() []operator.Operator {
	operators := dag.Operators()
	results := make([]operator.Operator, 0, len(operators))
	for _, op := range operators {
		if op.ExecState().HasWarning() {
			results = append(results, op)
		}
	}

	return results
}
