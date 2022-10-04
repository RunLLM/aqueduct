package handler

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/collections/operator_result"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_result"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
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
	workflowDagResultId uuid.UUID
	operatorId          uuid.UUID
}

type GetOperatorResultHandler struct {
	GetHandler

	Database                database.Database
	OperatorReader          operator.Reader
	OperatorResultReader    operator_result.Reader
	WorkflowDagResultReader workflow_dag_result.Reader
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
		aqContext.OrganizationId,
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
		AqContext:           aqContext,
		workflowDagResultId: workflowDagResultId,
		operatorId:          operatorId,
	}, http.StatusOK, nil
}

func (h *GetOperatorResultHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*getOperatorResultArgs)

	emptyResp := shared.ExecutionState{}

	dbWorkflowDagResult, err := h.WorkflowDagResultReader.GetWorkflowDagResult(
		ctx,
		args.workflowDagResultId,
		h.Database,
	)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred when retrieving workflow result.")
	}

	response := shared.ExecutionState{}
	dbOperatorResult, err := h.OperatorResultReader.GetOperatorResultByWorkflowDagResultIdAndOperatorId(
		ctx,
		args.workflowDagResultId,
		args.operatorId,
		h.Database,
	)
	if err != nil {
		if err != database.ErrNoRows {
			return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred when retrieving operator result.")
		}
		// OperatorResult was never created, so we use the WorkflowDagResult's status as this OperatorResult's status
		response.Status = dbWorkflowDagResult.Status
	} else {
		response.Status = dbOperatorResult.Status
	}

	if dbOperatorResult != nil && !dbOperatorResult.ExecState.IsNull {
		response.FailureType = dbOperatorResult.ExecState.FailureType
		response.Error = dbOperatorResult.ExecState.Error
		response.UserLogs = dbOperatorResult.ExecState.UserLogs
	}

	return response, http.StatusOK, nil
}
