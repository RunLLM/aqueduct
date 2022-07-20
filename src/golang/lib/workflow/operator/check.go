package operator

import (
	"context"
	"encoding/json"
	"fmt"

	db_artifact "github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/collections/operator/check"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/dropbox/godropbox/errors"
	log "github.com/sirupsen/logrus"
)

type checkOperatorImpl struct {
	baseFunctionOperator
}

func newCheckOperator(base baseFunctionOperator) (Operator, error) {
	base.jobName = generateFunctionJobName()

	inputs := base.inputs
	outputs := base.outputs
	if len(inputs) == 0 {
		return nil, errWrongNumInputs
	}
	if len(outputs) != 1 {
		return nil, errWrongNumOutputs
	}

	for _, inputArtifact := range inputs {
		if inputArtifact.Type() != db_artifact.TableType &&
			inputArtifact.Type() != db_artifact.FloatType &&
			inputArtifact.Type() != db_artifact.JsonType {
			return nil, errors.New("Inputs to metric operator must be Table, Float, or Parameter Artifacts.")
		}
	}
	for _, outputArtifact := range outputs {
		if outputArtifact.Type() != db_artifact.BoolType {
			return nil, errors.New("Outputs of function operator must be Table Artifacts.")
		}
	}

	return &checkOperatorImpl{
		base,
	}, nil
}

func (co *checkOperatorImpl) hasErrorSeverity() bool {
	return co.dbOperator.Spec.Check().Level == check.ErrorLevel
}

func (co *checkOperatorImpl) GetExecState(ctx context.Context) (*shared.ExecutionState, error) {
	execState, err := co.baseOperator.GetExecState(ctx)
	if err != nil {
		return nil, err
	}

	// If this was a blocking check that failed, overwrite the returned tip message
	// to be much more helpful.
	if co.hasErrorSeverity() &&
		execState.Status == shared.FailedExecutionStatus &&
		execState.Error.Tip == shared.TipBlacklistedOutputError {
		execState.Error.Tip = fmt.Sprintf("Check operator %s failed and has ERROR level severity, so the entire workflow failed.", co.Name())
	}
	return execState, nil
}

func (co *checkOperatorImpl) JobSpec() job.Spec {
	fn := co.dbOperator.Spec.Check().Function
	spec := co.jobSpec(&fn)

	// This will tell the orchestration engine to fail the workflow
	// if the check fails with sufficient severity.
	if co.hasErrorSeverity() {
		falseSerialized, err := json.Marshal(false)
		if err != nil {
			log.Errorf("Internal error: Operator %s is unable to marshal `false`")
		}

		fnSpec := spec.(*job.FunctionSpec)
		fnSpec.BlacklistedOutputs = append(fnSpec.BlacklistedOutputs, string(falseSerialized))
		return fnSpec
	}
	return spec
}
