package handler

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Route: /resource/{resourceID}/operators
// Method: GET
// Params: resourceID
// Request:
//	Headers:
//		`api-key`: user's API Key
// Response:
//	Body:
//		serialized `listOperatorsForResourceResponse`
//
// `listOperatorsForResource` lists all operators associated with
// the given resource. Together we provide the following information for
// each associated operator:
//  `workflow_id`: the workflow associated with this operator
//	`workflow_dag_id`: the workflow dag associated with this operator
//	`is_active`: whether the operator is being used in the latest version of the workflow.
//				 This is the equivalent to whether `workflow_dag_id` is the latest for the `workflow_id`

type listOperatorsForResourceItem struct {
	Operator      *models.Operator `json:"operator"`
	WorkflowId    uuid.UUID        `json:"workflow_id"`
	WorkflowDagId uuid.UUID        `json:"workflow_dag_id"`
	IsActive      bool             `json:"is_active"`
}

type listOperatorsForResourceArgs struct {
	*aq_context.AqContext
	resourceObject *models.Resource
}

type listOperatorsForResourceResponse struct {
	OperatorWithIds []listOperatorsForResourceItem `json:"operator_with_ids"`
}

type ListOperatorsResourecHandler struct {
	GetHandler

	Database database.Database

	DAGRepo      repos.DAG
	ResourceRepo repos.Resource
	OperatorRepo repos.Operator
}

func (*ListOperatorsResourecHandler) Name() string {
	return "ListOperatorsForResource"
}

func (h *ListOperatorsResourecHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statuscode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statuscode, err
	}

	resourceIDStr := chi.URLParam(r, routes.ResourceIDUrlParam)
	resourceID, err := uuid.Parse(resourceIDStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed resource ID.")
	}

	resourceObject, err := h.ResourceRepo.Get(r.Context(), resourceID, h.Database)
	if err != nil {
		return nil, http.StatusNotFound, errors.Wrap(err, "Failed to retrieve resource object.")
	}

	return &listOperatorsForResourceArgs{
		AqContext:      aqContext,
		resourceObject: resourceObject,
	}, http.StatusOK, nil
}

func (h *ListOperatorsResourecHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*listOperatorsForResourceArgs)

	operators, err := operator.GetOperatorsOnResource(
		ctx,
		args.OrgID,
		args.resourceObject,
		h.ResourceRepo,
		h.OperatorRepo,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to retrieve operators.")
	}
	if len(operators) == 0 {
		return listOperatorsForResourceResponse{OperatorWithIds: []listOperatorsForResourceItem{}}, http.StatusOK, nil
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
	results := make([]listOperatorsForResourceItem, 0, len(operatorRelations))
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
		results = append(results, listOperatorsForResourceItem{
			Operator:      &op,
			WorkflowId:    operatorRelation.WorkflowID,
			WorkflowDagId: operatorRelation.DagID,
			IsActive:      active,
		})
	}
	return listOperatorsForResourceResponse{OperatorWithIds: results}, http.StatusOK, nil
}
