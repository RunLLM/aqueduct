package operator

import (
	"context"
	"encoding/base64"
	"fmt"
	log "github.com/sirupsen/logrus"

	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/storage"
	"github.com/google/uuid"
)

func generateParamJobName() string {
	return fmt.Sprintf("param-operator-%s", uuid.New().String())
}

type paramOperatorImpl struct {
	baseOperator
}

func newParamOperator(
	base baseOperator,
) (Operator, error) {
	base.jobName = generateParamJobName()

	inputs := base.inputs
	outputs := base.outputs
	if len(inputs) != 0 {
		return nil, errWrongNumInputs
	}
	if len(outputs) != 1 {
		return nil, errWrongNumOutputs
	}

	// Write the parameter's value from the spec to the output content path.
	// This is to avoid passing raw values in the spec, which can be of unbounded
	// size (eg. images can be too large).
	// This also means that parameter artifacts are automatically considered `Computed`,
	// as soon they are constructed.
	paramValBytes, err := base64.StdEncoding.DecodeString(base.dbOperator.Spec.Param().Val)
	if err != nil {
		return nil, err
	}
	err = storage.NewStorage(base.storageConfig).Put(
		context.TODO(),
		base.outputExecPaths[0].ArtifactContentPath,
		paramValBytes,
	)
	if err != nil {
		return nil, err
	}

	return &paramOperatorImpl{
		base,
	}, nil
}

func (po *paramOperatorImpl) JobSpec() job.Spec {
	log.Errorf("HELLO: parameter serialization type: %s", po.dbOperator.Spec.Param().SerializationType)
	return &job.ParamSpec{
		BasePythonSpec: job.NewBasePythonSpec(
			job.ParamJobType,
			po.jobName,
			*po.storageConfig,
			po.metadataPath,
		),
		ExpectedType:      po.outputs[0].Type(),
		SerializationType: po.dbOperator.Spec.Param().SerializationType,

		OutputMetadataPath: po.outputExecPaths[0].ArtifactMetadataPath,
		OutputContentPath:  po.outputExecPaths[0].ArtifactContentPath,
	}
}

// All the parameter operator does is write the parameter value to storage,
// to be consuemd by downstream operators.
func (po *paramOperatorImpl) Launch(ctx context.Context) error {
	return po.launch(ctx, po.JobSpec())
}
