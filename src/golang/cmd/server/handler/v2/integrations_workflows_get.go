package v2

import (
	"context"
	"fmt"
	"github.com/aqueducthq/aqueduct/cmd/server/handler"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/dynamic"
	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/aqueducthq/aqueduct/lib/functional/slices"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/models/views"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator"
	"github.com/google/uuid"
	"net/http"
)

// This file should map directly to
// src/ui/common/src/handlers/v2/IntegrationsWorkflowsGet.ts
//
// Route: /v2/integrations/workflows
// Method: GET
// Request:
//	Headers:
//		`api-key`: user's API Key
// Response:
//	Body:
//		Map of integration ID to list of workflow IDs that use that integration.

type integrationsWorkflowsGetArgs struct {
	*aq_context.AqContext
}

type IntegrationsWorkflowsGetHandler struct {
	handler.GetHandler

	Database        database.Database
	IntegrationRepo repos.Integration
	OperatorRepo    repos.Operator
}

func (*IntegrationsWorkflowsGetHandler) Name() string {
	return "WorkflowsByIntegrationGet"
}

func (h *IntegrationsWorkflowsGetHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	return &integrationsWorkflowsGetArgs{
		AqContext: aqContext,
	}, http.StatusOK, nil
}

func (h *IntegrationsWorkflowsGetHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*integrationsWorkflowsGetArgs)

	integrations, err := h.IntegrationRepo.GetByUser(
		ctx,
		args.OrgID,
		args.ID,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to list integrations.")
	}

	response := make(map[uuid.UUID][]uuid.UUID, len(integrations))
	for _, integration := range integrations {
		workflowAndOperatorsList, err := h.fetchWorkflowIDsForIntegration(ctx, args.OrgID, &integration)
		if err != nil {
			return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to fetch workflow and operators for integration.")
		}
		response[integration.ID] = workflowAndOperatorsList
	}
	return response, http.StatusOK, nil
}

// fetchWorkflowIDsForIntegration returns a list of workflow IDs that use the given integration.
func (h *IntegrationsWorkflowsGetHandler) fetchWorkflowIDsForIntegration(
	ctx context.Context,
	orgID string,
	integration *models.Integration,
) ([]uuid.UUID, error) {
	integrationID := integration.ID

	// If the requested integration is a cloud integration, substitute the cloud integration ID
	// with the ID of the dynamic k8s integration.
	if integration.Service == shared.AWS {
		k8sIntegration, err := h.IntegrationRepo.GetByNameAndUser(
			ctx,
			fmt.Sprintf("%s:%s", integration.Name, dynamic.K8sIntegrationNameSuffix),
			uuid.Nil,
			orgID,
			h.Database,
		)
		if err != nil {
			return nil, err
		}

		integrationID = k8sIntegration.ID
	}

	operators, err := operator.GetOperatorsOnIntegration(
		ctx,
		integrationID,
		h.IntegrationRepo,
		h.OperatorRepo,
		h.Database,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to retrieve operators.")
	}

	// Now, using the operators using this integration, we can infer all the workflows
	// that also use this integration.
	operatorIDs := slices.Map(operators, func(op models.Operator) uuid.UUID {
		return op.ID
	})

	operatorRelations, err := h.OperatorRepo.GetRelationBatch(ctx, operatorIDs, h.Database)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to retrieve operator ID information.")
	}

	workflowIDs := slices.Map(operatorRelations, func(operatorRelation views.OperatorRelation) uuid.UUID {
		return operatorRelation.WorkflowID
	})
	return workflowIDs, nil
}
