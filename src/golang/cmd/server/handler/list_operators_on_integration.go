package handler

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/queries"
	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

// Route: /integration/{integrationId}/workflows
// Method: GET
// Params: None
// Request:
//	Headers:
//		`api-key`: user's API Key
// Response:
//	Body:
//		serialized `listOperatorsOnIntegrationResponse`
//
// `listOperatorsOnIntegration` lists all operators associated with
// the given integraion. Together we provide the following information for
// each associated operator:
//  `workflow_id`: the workflow associated with this operator
//	`workflow_dag_id`: the workflow dag associated with this operator
//	`is_active`: whether the operator is being used in the latest version of the workflow.
//				 This is the equivalent to whether `workflow_dag_id` is the latest for the `workflow_id`

type listOperatorsOnIntegrationItem struct {
	Operator      *operator.Operator
	WorkflowId    uuid.UUID
	WorkflowDagId uuid.UUID
	IsActive      bool `json:"is_active"`
}

type listOperatorsOnIntegrationResponse struct {
	OperatorWithIds []listOperatorsOnIntegrationItem `json:"operator_with_ids"`
}

type ListOperatorsOnIntegrationHandler struct {
	GetHandler

	Database       database.Database
	CustomReader   queries.Reader
	OperatorReader operator.Reader
}

func (*ListOperatorsOnIntegrationHandler) Name() string {
	return "ListWorkflowsOnIntegration"
}

func (*ListOperatorsOnIntegrationHandler) Prepare(r *http.Request) (interface{}, int, error) {
	integrationIdStr := chi.URLParam(r, routes.IntegrationIdUrlParam)
	integrationId, err := uuid.Parse(integrationIdStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed integration ID.")
	}

	return integrationId, http.StatusOK, nil
}

func (h *ListOperatorsOnIntegrationHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	integrationId := interfaceArgs.(uuid.UUID)
	operators, err := h.OperatorReader.GetOperatorsByIntegrationId(ctx, integrationId, h.Database)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to retrieve operators.")
	}

	operatorIds := make([]uuid.UUID, 0, len(operators))
	operatorByIds := make(map[uuid.UUID]operator.Operator, len(operators))
	for _, op := range operators {
		operatorIds = append(operatorIds, op.Id)
		operatorByIds[op.Id] = op
	}

	ids, err := h.CustomReader.GetWorkflowIdsFromOperatorIds(ctx, operatorIds, h.Database)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to retrieve operator ID information.")
	}

	workflowIds := make([]uuid.UUID, 0, len(ids))
	for _, idsItem := range ids {
		workflowIds = append(workflowIds, idsItem.WorkflowId)
	}

	workflowIdsToLatestDagIds, err := h.CustomReader.GetLatestWorkflowDagIdsFromWorkflowIds(
		ctx,
		workflowIds,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to retrieve latest dag IDs for workflows.")
	}

	results := make([]listOperatorsOnIntegrationItem, 0, len(ids))
	for _, idsItem := range ids {
		op, ok := operatorByIds[idsItem.OperatorId]
		if !ok {
			return nil, http.StatusInternalServerError, errors.New("Operator id mismatch retrieved operators.")
		}

		latestDagId, ok := workflowIdsToLatestDagIds[idsItem.WorkflowId]
		if !ok {
			return nil, http.StatusInternalServerError, errors.New("Workflow id mismatch retrieved workflows.")
		}

		active := latestDagId == idsItem.WorkflowDagId
		results = append(results, listOperatorsOnIntegrationItem{
			Operator:      &op,
			WorkflowId:    idsItem.WorkflowId,
			WorkflowDagId: idsItem.WorkflowDagId,
			IsActive:      active,
		})
	}

	return listOperatorsOnIntegrationResponse{OperatorWithIds: results}, http.StatusOK, nil
}
