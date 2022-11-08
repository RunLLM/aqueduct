package environment

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	"github.com/aqueducthq/aqueduct/lib/database"
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

func (e *Environment) CondaName() string {
	return fmt.Sprintf("aqueduct_%s", e.Id.String())
}

// `CreateConda` creates the conda environment based on this
// environment's python version and dependencies.
func (e *Environment) CreateConda() error {
	return nil
}

func (e *Environment) DeleteCondaIfExists() error {
	return nil
}
