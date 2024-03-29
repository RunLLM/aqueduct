package v2

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/handler"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/aqueducthq/aqueduct/lib/functional/slices"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/response"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator"
	"github.com/google/uuid"
)

// This file should map directly to
// src/ui/common/src/handlers/v2/ResourcesWorkflowsGet.ts
//
// Route: /v2/resources/workflows
// Method: GET
// Request:
//	Headers:
//		`api-key`: user's API Key
// Response:
//	Body:
//		Map of resource ID to list of `response.WorkflowAndDagID` that use that resource.

type resourcesWorkflowsGetArgs struct {
	*aq_context.AqContext
}

type ResourcesWorkflowsGetHandler struct {
	handler.GetHandler

	Database      database.Database
	ResourceRepo  repos.Resource
	WorkflowRepo  repos.Workflow
	DAGRepo       repos.DAG
	DAGResultRepo repos.DAGResult
	OperatorRepo  repos.Operator
}

func (*ResourcesWorkflowsGetHandler) Name() string {
	return "ResourcesWorkflowsGet"
}

func (h *ResourcesWorkflowsGetHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	return &resourcesWorkflowsGetArgs{
		AqContext: aqContext,
	}, http.StatusOK, nil
}

func (h *ResourcesWorkflowsGetHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*resourcesWorkflowsGetArgs)

	resources, err := h.ResourceRepo.GetByUser(
		ctx,
		args.OrgID,
		args.ID,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to list resources.")
	}

	resp := make(map[uuid.UUID][]*response.WorkflowAndDagIDs, len(resources))
	for _, resource := range resources {
		workflowAndDagIDs, err := fetchWorkflowAndDagIDsForResource(
			ctx,
			args.OrgID, &resource, h.ResourceRepo, h.WorkflowRepo, h.OperatorRepo, h.DAGRepo, h.DAGResultRepo, h.Database)
		if err != nil {
			return nil, http.StatusInternalServerError, errors.Wrapf(err, "Unable to find workflows for resource %s", resource.ID)
		}
		resp[resource.ID] = workflowAndDagIDs
	}
	return resp, http.StatusOK, nil
}

// fetchWorkflowAndDagIDsForResource returns a list of workflow IDs that use the given resource.
// We consider a workflow to use a resource if it has run an operator that uses this resource during
// it's latest run.
func fetchWorkflowAndDagIDsForResource(
	ctx context.Context,
	orgID string,
	resource *models.Resource,
	resourceRepo repos.Resource,
	workflowRepo repos.Workflow,
	operatorRepo repos.Operator,
	dagRepo repos.DAG,
	dagResultRepo repos.DAGResult,
	db database.Database,
) ([]*response.WorkflowAndDagIDs, error) {
	// For performance reasons, we split out the workflows fetching for notifications, since for these
	// resources, you can fetch the workflow IDs that use them directly, instead of having to go through
	// operators.
	if shared.IsNotificationResource(resource.Service) {
		workflowIDs, err := operator.GetWorkflowIDsUsingNotification(ctx, resource, workflowRepo, db)
		if err != nil {
			return nil, err
		}

		workflowIDToLatestDagID, err := dagRepo.GetLatestIDByWorkflowBatch(ctx, workflowIDs, db)
		if err != nil {
			return nil, err
		}

		workflowAndDagIDs := make([]*response.WorkflowAndDagIDs, 0, len(workflowIDToLatestDagID))
		for workflowID, dagID := range workflowIDToLatestDagID {
			workflowAndDagIDs = append(workflowAndDagIDs, &response.WorkflowAndDagIDs{
				WorkflowID: workflowID,
				DagID:      dagID,
			})
		}
		return workflowAndDagIDs, nil

	} else {
		operators, err := operator.GetOperatorsOnResource(
			ctx,
			orgID,
			resource,
			resourceRepo,
			operatorRepo,
			db,
		)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to retrieve operators.")
		}

		// Now, using the operators using this resource, we can infer all the workflows
		// that also use this resource.
		operatorIDs := slices.Map(operators, func(op models.Operator) uuid.UUID {
			return op.ID
		})

		operatorRelations, err := operatorRepo.GetRelationBatch(ctx, operatorIDs, db)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to retrieve operator ID information.")
		}

		// This map is derived directly from the operators.
		workflowIDToDagIDs := make(map[uuid.UUID][]uuid.UUID, len(operatorRelations))
		for _, operatorRelation := range operatorRelations {
			workflowIDToDagIDs[operatorRelation.WorkflowID] = append(
				workflowIDToDagIDs[operatorRelation.WorkflowID],
				operatorRelation.DagID,
			)
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
}
