package handler

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Route: /integration/{integrationId}/operators
// Method: GET
// Params: integrationId
// Request:
//	Headers:
//		`api-key`: user's API Key
// Response:
//	Body:
//		serialized `listOperatorsForIntegrationResponse`
//
// `listOperatorsForIntegration` lists all operators associated with
// the given integraion. Together we provide the following information for
// each associated operator:
//  `workflow_id`: the workflow associated with this operator
//	`workflow_dag_id`: the workflow dag associated with this operator
//	`is_active`: whether the operator is being used in the latest version of the workflow.
//				 This is the equivalent to whether `workflow_dag_id` is the latest for the `workflow_id`

type listOperatorsForIntegrationItem struct {
	Operator      *models.Operator `json:"operator"`
	WorkflowId    uuid.UUID        `json:"workflow_id"`
	WorkflowDagId uuid.UUID        `json:"workflow_dag_id"`
	IsActive      bool             `json:"is_active"`
}

type listOperatorsForIntegrationResponse struct {
	OperatorWithIds []listOperatorsForIntegrationItem `json:"operator_with_ids"`
}

type ListOperatorsForIntegrationHandler struct {
	GetHandler

	Database database.Database

	DAGRepo         repos.DAG
	IntegrationRepo repos.Integration
	OperatorRepo    repos.Operator
}

func (*ListOperatorsForIntegrationHandler) Name() string {
	return "ListOperatorsForIntegration"
}

func (*ListOperatorsForIntegrationHandler) Prepare(r *http.Request) (interface{}, int, error) {
	integrationIDStr := chi.URLParam(r, routes.IntegrationIdUrlParam)
	integrationID, err := uuid.Parse(integrationIDStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed integration ID.")
	}

	return integrationID, http.StatusOK, nil
}

func (h *ListOperatorsForIntegrationHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	integrationID := interfaceArgs.(uuid.UUID)
	operators, err := operator.GetOperatorsOnIntegration(
		ctx,
		integrationID,
		h.IntegrationRepo,
		h.OperatorRepo,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to retrieve operators.")
	}

	operatorIDs := make([]uuid.UUID, 0, len(operators))
	operatorByIDs := make(map[uuid.UUID]models.Operator, len(operators))
	for _, op := range operators {
		operatorIDs = append(operatorIDs, op.ID)
		operatorByIDs[op.ID] = op
	}

	// Fetch all workflows that owns all fetched operators.
	operatorRelations, err := h.OperatorRepo.GetRelationBatch(ctx, operatorIDs, h.Database)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to retrieve operator ID information.")
	}

	workflowIDs := make([]uuid.UUID, 0, len(operatorRelations))
	for _, operatorRelation := range operatorRelations {
		workflowIDs = append(workflowIDs, operatorRelation.WorkflowID)
	}

	// Fetch latest workflow_dag_id for each workflow.
	workflowIdsToLatestDagIds, err := h.DAGRepo.GetLatestIDByWorkflowBatch(
		ctx,
		workflowIDs,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to retrieve latest dag IDs for workflows.")
	}

	// Combine all fetched results
	results := make([]listOperatorsForIntegrationItem, 0, len(operatorRelations))
	for _, operatorRelation := range operatorRelations {
		op, ok := operatorByIDs[operatorRelation.OperatorID]
		if !ok {
			return nil, http.StatusInternalServerError, errors.New("Operator id mismatch retrieved operators.")
		}

		latestDagId, ok := workflowIdsToLatestDagIds[operatorRelation.WorkflowID]
		if !ok {
			return nil, http.StatusInternalServerError, errors.New("Workflow id mismatch retrieved workflows.")
		}

		active := latestDagId == operatorRelation.DagID
		results = append(results, listOperatorsForIntegrationItem{
			Operator:      &op,
			WorkflowId:    operatorRelation.WorkflowID,
			WorkflowDagId: operatorRelation.DagID,
			IsActive:      active,
		})
	}
	return listOperatorsForIntegrationResponse{OperatorWithIds: results}, http.StatusOK, nil
}
