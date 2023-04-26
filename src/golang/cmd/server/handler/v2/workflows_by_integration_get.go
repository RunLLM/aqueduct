package v2

import (
	"context"
	"fmt"
	"github.com/aqueducthq/aqueduct/cmd/server/handler"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/dynamic"
	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator"
	"github.com/google/uuid"
	"net/http"
)

// This file should map directly to
// src/ui/common/src/handlers/v2/WorkflowsByIntegrationGet.tsx
//
// Route: /v2/workflows_by_integration
// Method: GET
// Request:
//	Headers:
//		`api-key`: user's API Key
// Response:
//	Body:
//		Map of integration ID to list of `workflowAndOperators` that use it.

type workflowsByIntegrationGetArgs struct {
	*aq_context.AqContext
}

type WorkflowsByIntegrationGetHandler struct {
	handler.GetHandler

	Database        database.Database
	IntegrationRepo repos.Integration
	WorkflowRepo    repos.Workflow
	OperatorRepo    repos.Operator
	DAGRepo         repos.DAG
}

// workflowAndOperators represents a single workflow and it's operators.
// This is used as the values in a map keyed by integration ID. The operators
// in this struct are only those in the workflow that use the given integration.
type workflowAndOperators struct {
	workflowID uuid.UUID
	operators  []models.Operator
}

func (*WorkflowsByIntegrationGetHandler) Name() string {
	return "WorkflowsByIntegrationGet"
}

func (h *WorkflowsByIntegrationGetHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	return &workflowsByIntegrationGetArgs{
		AqContext: aqContext,
	}, http.StatusOK, nil
}

func (h *WorkflowsByIntegrationGetHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*workflowsByIntegrationGetArgs)

	integrations, err := h.IntegrationRepo.GetByUser(
		ctx,
		args.OrgID,
		args.ID,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to list integrations.")
	}

	response := make(map[uuid.UUID][]*workflowAndOperators, len(integrations))
	for _, integration := range integrations {
		workflowAndOperatorsList, err := h.fetchWorkflowAndOperatorsForIntegration(ctx, args.OrgID, &integration)
		if err != nil {
			return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to fetch workflow and operators for integration.")
		}
		response[integration.ID] = workflowAndOperatorsList
	}
	return response, http.StatusOK, nil
}

// fetchWorkflowAndOperatorsForIntegration returns a list of workflowAndOperators
// objects for the given integration.
func (h *WorkflowsByIntegrationGetHandler) fetchWorkflowAndOperatorsForIntegration(
	ctx context.Context,
	orgID string,
	integration *models.Integration,
) ([]*workflowAndOperators, error) {
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
	operatorsByWorkflowID := make(map[uuid.UUID][]models.Operator, 1)

	operatorIDs := make([]uuid.UUID, 0, len(operators))
	operatorByIDs := make(map[uuid.UUID]models.Operator, len(operators))
	for _, op := range operators {
		operatorIDs = append(operatorIDs, op.ID)
		operatorByIDs[op.ID] = op
	}
	operatorRelations, err := h.OperatorRepo.GetRelationBatch(ctx, operatorIDs, h.Database)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to retrieve operator ID information.")
	}
	for _, operatorRelation := range operatorRelations {
		workflowID := operatorRelation.WorkflowID
		if _, ok := operatorsByWorkflowID[workflowID]; !ok {
			operatorsByWorkflowID[workflowID] = make([]models.Operator, 0, 1)
		}

		operatorsByWorkflowID[workflowID] = append(
			operatorsByWorkflowID[workflowID],
			operatorByIDs[operatorRelation.OperatorID],
		)
	}

	// Convert the workflow -> operators map into a list of workflowAndOperators.
	workflowAndOperatorsList := make([]*workflowAndOperators, len(operatorsByWorkflowID))
	for workflowID, operators := range operatorsByWorkflowID {
		workflowAndOperatorsList = append(
			workflowAndOperatorsList,
			&workflowAndOperators{
				workflowID: workflowID,
				operators:  operators,
			},
		)
	}

	return workflowAndOperatorsList, nil
}
