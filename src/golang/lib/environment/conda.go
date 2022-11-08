package environment

import (
	"context"

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
