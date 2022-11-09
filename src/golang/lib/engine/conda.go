package engine

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/environment"
	"github.com/google/uuid"
)

func IsCondaConnected(
	ctx context.Context,
	userId uuid.UUID,
	integrationReader integration.Reader,
	db database.Database,
) (bool, error) {
	integrations, err := integrationReader.GetIntegrationsByServiceAndUser(
		ctx,
		integration.Conda,
		userId,
		db,
	)
	if err != nil {
		return false, err
	}

	return len(integrations) > 0, nil
}

func CondaName(env *environment.Environment) string {
	return fmt.Sprintf("aqueduct_%s", env.Id.String())
}

// `CreateConda` creates the conda environment based on this
// environment's python version and dependencies.
func CreateConda(env *environment.Environment) error {
	return nil
}

func DeleteCondaIfExists(env *environment.Environment) error {
	return nil
}
