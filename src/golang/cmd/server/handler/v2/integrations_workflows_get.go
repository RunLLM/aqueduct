package v2

import (
	"context"
	"github.com/aqueducthq/aqueduct/lib/response"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/handler"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/aqueducthq/aqueduct/lib/functional/slices"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator"
	"github.com/google/uuid"
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
//		Map of integration ID to list of `response.WorkflowAndDagID` that use that integration.

type integrationsWorkflowsGetArgs struct {
	*aq_context.AqContext
}

type IntegrationsWorkflowsGetHandler struct {
	handler.GetHandler

	Database        database.Database
	IntegrationRepo repos.Integration
	OperatorRepo    repos.Operator
	DAGResultRepo   repos.DAGResult
}

func (*IntegrationsWorkflowsGetHandler) Name() string {
	return "IntegrationsWorkflowsGet"
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

	response := make(map[uuid.UUID][]*response.WorkflowAndDagIDs, len(integrations))
	for _, integration := range integrations {
		workflowAndDagIDs, err := fetchWorkflowAndDagIDsForIntegration(ctx, args.OrgID, &integration, h.IntegrationRepo, h.OperatorRepo, h.DAGResultRepo, h.Database)
		if err != nil {
			return nil, http.StatusInternalServerError, errors.Wrapf(err, "Unable to find workflows for integration %s", integration.ID)
		}
		response[integration.ID] = workflowAndDagIDs
	}
	return response, http.StatusOK, nil
}

// fetchWorkflowAndDagIDsForIntegration returns a list of workflow IDs that use the given integration.
// We consider a workflow to use a resource if it has run an operator that uses this resource during
// it's latest run.
func fetchWorkflowAndDagIDsForIntegration(
	ctx context.Context,
	orgID string,
	integration *models.Integration,
	integrationRepo repos.Integration,
	operatorRepo repos.Operator,
	dagResultRepo repos.DAGResult,
	db database.Database,
) ([]*response.WorkflowAndDagIDs, error) {
	operators, err := operator.GetOperatorsOnIntegration(
		ctx,
		orgID,
		integration,
		integrationRepo,
		operatorRepo,
		db,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to retrieve operators.")
	}

	// Now, using the operators using this integration, we can infer all the workflows
	// that also use this integration.
	operatorIDs := slices.Map(operators, func(op models.Operator) uuid.UUID {
		return op.ID
	})

	operatorRelations, err := operatorRepo.GetRelationBatch(ctx, operatorIDs, db)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to retrieve operator ID information.")
	}

	// This map is derived directly from the operators.
	workflowIDToDagIDs := make(map[uuid.UUID][]uuid.UUID)
	for _, operatorRelation := range operatorRelations {
		if _, ok := workflowIDToDagIDs[operatorRelation.WorkflowID]; ok {
			workflowIDToDagIDs[operatorRelation.WorkflowID] = append(
				workflowIDToDagIDs[operatorRelation.WorkflowID],
				operatorRelation.DagID,
			)
		} else {
			workflowIDToDagIDs[operatorRelation.WorkflowID] = []uuid.UUID{operatorRelation.DagID}
		}
	}

	// For each workflow, fetch the latest dag ID. We can use this latest dag ID to filter out any
	// workflows had historically used this resource, but no longer do in their latest run.
	workflowAndDagIDs := make([]*response.WorkflowAndDagIDs, 0, len(workflowIDToDagIDs))
	for workflowID, dagIDs := range workflowIDToDagIDs {
		dbDAGResults, err := dagResultRepo.GetByWorkflow(ctx, workflowID, "created_at", 1, true, db)
		if err != nil {
			return nil, err
		}

		// Skip any workflows that have been defined but have not run yet.
		if len(dbDAGResults) == 1 {
			latestDagID := dbDAGResults[0].DagID

			found := false
			for _, dagID := range dagIDs {
				if dagID == latestDagID {
					found = true
					break
				}
			}

			// If the latest dag does not have any of the resource operator's on it, that means
			// that the workflow no longer uses this resource.
			if found {
				workflowAndDagIDs = append(workflowAndDagIDs, &response.WorkflowAndDagIDs{
					WorkflowID: workflowID,
					DagID:      latestDagID,
				})
			}
		}
	}

	return workflowAndDagIDs, nil
}
