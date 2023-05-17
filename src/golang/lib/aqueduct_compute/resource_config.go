package aqueduct_compute

import (
	"context"
	"encoding/json"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/execution_environment"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/google/uuid"
)

const (
	// Only set if the conda is registered. Represents the json-serialized string of the Conda resource config.
	CondaConfigKey     = "conda_config_serialized"
	CondaResourceIDKey = "conda_resource_id"
	CondaNameIDKey     = "conda_resource_name"

	// Only set if the conda is not registered.
	PythonVersionKey = "python_version"
)

// ConstructAqueductComputeResourceConfig constructs the config for the Aqueduct compute resource.
// If a conda resource has been registered, we embed it within the Aqueduct compute resource.
// Otherwise, we perform a best-effort fetch of the server's python version.
func ConstructAqueductComputeResourceConfig(
	ctx context.Context,
	userID uuid.UUID,
	integrationRepo repos.Resource,
	db database.Database,
) (shared.ResourceConfig, error) {
	condaResource, err := execution_environment.GetCondaResource(ctx, userID, integrationRepo, db)
	if err != nil {
		return nil, err
	}

	config := make(shared.ResourceConfig, 1)
	if condaResource != nil {
		condaSerialized, err := json.Marshal(condaResource.Config)
		if err != nil {
			return nil, err
		}
		config[CondaConfigKey] = string(condaSerialized)
		config[CondaResourceIDKey] = condaResource.ID.String()
		config[CondaNameIDKey] = condaResource.Name

	} else {
		config[PythonVersionKey] = execution_environment.GetServerPythonVersion()
	}
	return config, nil
}
