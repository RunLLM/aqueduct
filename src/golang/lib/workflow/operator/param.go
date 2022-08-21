package operator

import (
	"context"
	"fmt"
	"encoding/json"

	"github.com/aqueducthq/aqueduct/lib/storage"
	"github.com/aqueducthq/aqueduct/lib/job"
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

	return &paramOperatorImpl{
		base,
	}, nil
}

func (po *paramOperatorImpl) JobSpec() job.Spec {
	return &job.ParamSpec{
		BasePythonSpec: job.NewBasePythonSpec(
			job.ParamJobType,
			po.jobName,
			*po.storageConfig,
			po.metadataPath,
		),
		InputMetadataPath:  po.outputExecPaths[0].ArtifactMetadataPath,
		Val:                po.dbOperator.Spec.Param().Val,
		Val_Type:			po.dbOperator.Spec.Param().Val_Type,
		OutputMetadataPath: po.outputExecPaths[0].ArtifactMetadataPath,
		OutputContentPath:  po.outputExecPaths[0].ArtifactContentPath,
	}
}

func (po *paramOperatorImpl) Launch(ctx context.Context) error {
	param_metadata := make(map[string]string)
	system_metadata := make(map[string]string)
	system_metadata["val"] = po.dbOperator.Spec.Param().Val
	param_metadata["system_metadata"] = fmt.Sprint(system_metadata)
	byte_metadata, err := json.Marshal(param_metadata)
	storages := storage.NewStorage(po.storageConfig)
	err2 := storages.Put(ctx, po.outputExecPaths[0].ArtifactMetadataPath, byte_metadata)
	return po.launch(ctx, po.JobSpec())
}
