package server

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/utils"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/collections/operator_result"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

// Route: /operator_result/{workflowDagResultId}/{operatorId}
// Method: GET
// Params:
//	`workflowDagResultId`: ID for `workflow_dag_result` object
//	`operatorId`: ID for `operator` object
// Request:
//	Headers:
//		`api-key`: user's API Key
// Response:
//	Body:
//		serialized `getOperatorResultResponse`,
//		metadata and content of the result of `operatorId` on the given workflow_dag_result object.
type getOperatorResultArgs struct {
	*CommonArgs
	workflowDagResultId uuid.UUID
	operatorId          uuid.UUID
}

type getOperatorResultResponse struct {
	Status shared.ExecutionStatus `json:"status"`
	Error  string                 `json:"error"`
	Logs   map[string]string      `json:"logs"`
}

type GetOperatorResultHandler struct {
	GetHandler

	Database             database.Database
	OperatorReader       operator.Reader
	OperatorResultReader operator_result.Reader
}

func (*GetOperatorResultHandler) Name() string {
	return "GetOperatorResult"
}

func (h *GetOperatorResultHandler) Prepare(r *http.Request) (interface{}, int, error) {
	common, statusCode, err := ParseCommonArgs(r)
	if err != nil {
		return nil, statusCode, err
	}

	workflowDagResultIdStr := chi.URLParam(r, utils.WorkflowDagResultIdUrlParam)
	workflowDagResultId, err := uuid.Parse(workflowDagResultIdStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed workflow dag result ID.")
	}

	operatorIdStr := chi.URLParam(r, utils.OperatorIdUrlParam)
	operatorId, err := uuid.Parse(operatorIdStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed operator ID.")
	}

	ok, err := h.OperatorReader.ValidateOperatorOwnership(
		r.Context(),
		common.OrganizationId,
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
		CommonArgs:          common,
		workflowDagResultId: workflowDagResultId,
		operatorId:          operatorId,
	}, http.StatusOK, nil
}

func (h *GetOperatorResultHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*getOperatorResultArgs)

	emptyResp := getOperatorResultResponse{}

	dbOperatorResult, err := h.OperatorResultReader.GetOperatorResultByWorkflowDagResultIdAndOperatorId(
		ctx,
		args.workflowDagResultId,
		args.operatorId,
		h.Database,
	)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error occurred when retrieving operator result.")
	}

	response := getOperatorResultResponse{
		Status: dbOperatorResult.Status,
	}

	if !dbOperatorResult.Metadata.IsNull {
		response.Error = dbOperatorResult.Metadata.Error
		response.Logs = dbOperatorResult.Metadata.Logs
	}

	return response, http.StatusOK, nil
}
