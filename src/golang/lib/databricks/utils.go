package databricks

import (
	"context"

	"github.com/aqueducthq/aqueduct/config"
	"github.com/aqueducthq/aqueduct/lib/storage"
	"github.com/dropbox/godropbox/errors"
)

func AddEntrypointFilesToStorage(ctx context.Context) error {
	config := config.Storage()
	storageManager := storage.NewStorage(&config)

	filesToWrite := map[string]string{
		DatabricksFunctionScript: FunctionEntrypoint,
		DatabricksDataScript:     DataEntrypoint,
		DatabricksMetricScript:   SystemMetricEntrypoint,
		DatabricksParamScript:    ParamEntrypoint,
	}

	for fileName, fileContent := range filesToWrite {
		err := storageManager.Put(ctx, fileName, []byte(fileContent))
		if err != nil {
			return errors.Wrap(err, "Unable to upload Databricks entrypoint script to storage.")
		}
	}
	return nil
}
