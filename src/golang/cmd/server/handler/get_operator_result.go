package handler

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Route: /operator/{workflowDagResultId}/{operatorId}/result
// Method: GET
// Params:
//
//	`workflowDagResultId`: ID for `workflow_dag_result` object
//	`operatorId`: ID for `operator` object
//
// Request:
//
//	Headers:
//		`api-key`: user's API Key
//
// Response:
//
//	Body:
//		serialized `getOperatorResultResponse`,
//		metadata and content of the result of `operatorId` on the given workflow_dag_result object.
type getOperatorResultArgs struct {
	*aq_context.AqContext
	dagResultID uuid.UUID
	operatorID  uuid.UUID
}

type GetOperatorResultHandler struct {
	GetHandler

	Database       database.Database
	OperatorReader operator.Reader

	DAGResultRepo      repos.DAGResult
	OperatorResultRepo repos.OperatorResult
}

type GetOperatorResultResponse struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	ExecState   shared.ExecutionState  `json:"exec_state"`
	Status      shared.ExecutionStatus `json:"status"`
}

func (*GetOperatorResultHandler) Name() string {
	return "GetOperatorResult"
}

func (h *GetOperatorResultHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	workflowDagResultIdStr := chi.URLParam(r, routes.WorkflowDagResultIdUrlParam)
	workflowDagResultId, err := uuid.Parse(workflowDagResultIdStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed workflow dag result ID.")
	}

	operatorIdStr := chi.URLParam(r, routes.OperatorIdUrlParam)
	operatorId, err := uuid.Parse(operatorIdStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed operator ID.")
	}

	ok, err := h.OperatorReader.ValidateOperatorOwnership(
		r.Context(),
		aqContext.OrgID,
		operatorId,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error during operator ownership validation.")
	}
	if !ok {
		return nil, http.StatusBadRequest, errors.Wrap(err, "The organization does not own this operator.")
	}

	return &getOperatorResultArgs{
		AqContext:   aqContext,
		dagResultID: workflowDagResultId,
		operatorID:  operatorId,
	}, http.StatusOK, nil
}

func (h *GetOperatorResultHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*getOperatorResultArgs)

	emptyResp := GetOperatorResultResponse{}
	dbOperator, err := h.OperatorReader.GetOperator(ctx, args.operatorID, h.Database)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred when retrieving operator.")
	}

	dagResult, err := h.DAGResultRepo.Get(ctx, args.dagResultID, h.Database)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred when retrieving workflow result.")
	}

	dbOperatorResult, err := h.OperatorResultRepo.GetByDAGResultAndOperator(
		ctx,
		args.dagResultID,
		args.operatorID,
		h.Database,
	)

	executionState := shared.ExecutionState{
		Status: dbOperatorResult.ExecState.Status,
	}

	if err != nil {
		if err != database.ErrNoRows {
			return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred when retrieving operator result.")
		}
		// OperatorResult was never created, so we use the WorkflowDagResult's status as this OperatorResult's status
		executionState.Status = shared.ExecutionStatus(dagResult.Status)
	}

	if dbOperatorResult != nil && !dbOperatorResult.ExecState.IsNull {
		executionState.FailureType = dbOperatorResult.ExecState.FailureType
		executionState.Error = dbOperatorResult.ExecState.Error
		executionState.UserLogs = dbOperatorResult.ExecState.UserLogs
		executionState.Status = dbOperatorResult.ExecState.Status
	}

	response := GetOperatorResultResponse{
		ExecState: executionState, Name: dbOperator.Name, Description: dbOperator.Description, Status: executionState.Status,
	}
	return response, http.StatusOK, nil
}
