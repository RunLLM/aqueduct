package operator

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared/operator"
	"github.com/aqueducthq/aqueduct/lib/storage"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

// UploadOperatorFiles uploads the provided Operator files to storage for the WorkflowDag specified.
// It updates the relevant operator spec with the storage path. It returns an error, if any.
func UploadOperatorFiles(
	ctx context.Context,
	dag *models.DAG,
	operatorIdToFileContents map[uuid.UUID][]byte,
) ([]string, error) {
	paths := make([]string, 0, len(operatorIdToFileContents))

	for operatorId, content := range operatorIdToFileContents {
		path := fmt.Sprintf("operator-%s", uuid.New())
		paths = append(paths, path)
		if err := storage.NewStorage(&dag.StorageConfig).Put(ctx, path, content); err != nil {
			return paths, err
		}

		operatorObject, ok := dag.Operators[operatorId]
		if !ok {
			return paths, errors.Newf("Unable to find operator %v in DAG operators.", operatorId)
		}

		if err := updateOperatorSpecFilePath(&operatorObject.Spec, path); err != nil {
			return paths, err
		}
	}

	return paths, nil
}

func updateOperatorSpecFilePath(spec *operator.Spec, filePath string) error {
	switch {
	case spec.IsFunction():
		spec.Function().StoragePath = filePath
	case spec.IsMetric():
		spec.Metric().Function.StoragePath = filePath
	case spec.IsCheck():
		spec.Check().Function.StoragePath = filePath
	default:
		return errors.Newf("Storage file path can only be set for a Function spec.")
	}
	return nil
}
