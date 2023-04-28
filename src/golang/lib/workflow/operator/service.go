package operator

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/dynamic"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

func GetOperatorsOnIntegration(
	ctx context.Context,
	orgID string,
	integration *models.Integration,
	integrationRepo repos.Integration,
	operatorRepo repos.Operator,
	DB database.Database,
) ([]models.Operator, error) {
	integrationID := integration.ID

	// If the requested integration is a cloud integration, substitute the cloud integration ID
	// with the ID of the dynamic k8s integration.
	if integration.Service == shared.AWS {
		k8sIntegration, err := integrationRepo.GetByNameAndUser(
			ctx,
			fmt.Sprintf("%s:%s", integration.Name, dynamic.K8sIntegrationNameSuffix),
			uuid.Nil,
			orgID,
			DB,
		)
		if err != nil {
			return nil, err
		}

		integrationID = k8sIntegration.ID
	}

	integrationObject, err := integrationRepo.Get(ctx, integrationID, DB)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to retrieve integration.")
	}

	if shared.IsDataIntegration(integrationObject.Service) {
		return operatorRepo.GetExtractAndLoadOPsByIntegration(ctx, integrationID, DB)
	}

	if _, ok := shared.ServiceToEngineConfigField[integrationObject.Service]; ok {
		return operatorRepo.GetByEngineIntegrationID(ctx, integrationID, DB)
	}

	// Other eligible cases
	if integrationObject.Service == shared.Conda {
		return operatorRepo.GetByEngineType(ctx, shared.AqueductCondaEngineType, DB)
	}

	// This feature is not supported for the given service.
	return nil, nil
}
