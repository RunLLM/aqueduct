package operator

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

func GetOperatorsOnIntegration(
	ctx context.Context,
	integrationID uuid.UUID,
	integrationRepo repos.Integration,
	operatorRepo repos.Operator,
	DB database.Database,
) ([]models.Operator, error) {
	integrationObject, err := integrationRepo.Get(ctx, integrationID, DB)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to retrieve integration.")
	}

	if shared.IsDatabaseIntegration(integrationObject.Service) {
		return operatorRepo.GetExtractAndLoadOPsByIntegration(ctx, integrationID, DB)
	}

	if _, ok := shared.ServiceToEngineConfigIntegrationIDField[integrationObject.Service]; ok {
		return operatorRepo.GetByEngineIntegrationID(ctx, integrationID, DB)
	}

	// Other eligible cases
	if integrationObject.Service == shared.Conda {
		return operatorRepo.GetByEngineType(ctx, shared.AqueductCondaEngineType, DB)
	}

	// This feature is not supported for the given service.
	return nil, nil
}
