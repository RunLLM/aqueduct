package handler

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/request"
	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/engine"
	shared_utils "github.com/aqueducthq/aqueduct/lib/lib_utils"
	"github.com/aqueducthq/aqueduct/lib/models/shared/operator/param"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type RefreshWorkflowArgs struct {
	WorkflowId uuid.UUID
	Parameters map[string]param.Param
}

// Route: /workflow/{workflowId}/refresh
// Method: POST
// Params: workflowId
// Request:
//
//	Headers:
//		`api-key`: user's API Key
//
// Response: none
//
// Refresh workflow creates a new workflow version by
// triggering running a workflow run.
type RefreshWorkflowHandler struct {
	PostHandler

	Database database.Database
	Engine   engine.Engine
}

func (*RefreshWorkflowHandler) Name() string {
	return "RefreshWorkflow"
}

func (h *RefreshWorkflowHandler) Prepare(r *http.Request) (interface{}, int, error) {
	workflowIDStr := chi.URLParam(r, routes.WorkflowIdUrlParam)
	if workflowIDStr == "" {
		return nil, http.StatusBadRequest, errors.New("no workflow id was specified")
	}

	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed workflow ID.")
	}

	parameters, err := request.ExtractParamsfromRequest(r)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "The user-defined parameters could not be extracted in current format.")
	}

	return &RefreshWorkflowArgs{
		WorkflowId: workflowID,
		Parameters: parameters,
	}, http.StatusOK, nil
}

func (h *RefreshWorkflowHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*RefreshWorkflowArgs)

	emptyResp := struct{}{}

	timeConfig := &engine.AqueductTimeConfig{
		OperatorPollInterval: engine.DefaultPollIntervalMillisec,
		ExecTimeout:          engine.DefaultExecutionTimeout,
		CleanupTimeout:       engine.DefaultCleanupTimeout,
	}

	_, err := h.Engine.TriggerWorkflow(
		ctx,
		args.WorkflowId,
		shared_utils.AppendPrefix(args.WorkflowId.String()),
		timeConfig,
		args.Parameters,
	)
	if err != nil {
		return emptyResp, http.StatusInternalServerError, errors.Wrap(err, "Unable to trigger workflow.")
	}

	return emptyResp, http.StatusOK, nil
}
