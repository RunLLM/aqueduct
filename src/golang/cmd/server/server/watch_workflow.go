package server

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/utils"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_watcher"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type watchWorkflowArgs struct {
	*CommonArgs
	workflowId uuid.UUID
}

type WatchWorkflowHandler struct {
	PostHandler

	Database              database.Database
	WorkflowReader        workflow.Reader
	WorkflowWatcherWriter workflow_watcher.Writer
}

func (*WatchWorkflowHandler) Name() string {
	return "WatchWorkflow"
}

func (h *WatchWorkflowHandler) Prepare(r *http.Request) (interface{}, int, error) {
	common, statusCode, err := ParseCommonArgs(r)
	if err != nil {
		return nil, statusCode, err
	}

	workflowIdStr := chi.URLParam(r, utils.WorkflowIdUrlParam)
	workflowId, err := uuid.Parse(workflowIdStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed workflow ID.")
	}

	ok, err := h.WorkflowReader.ValidateWorkflowOwnership(
		r.Context(),
		workflowId,
		common.OrganizationId,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error during workflow ownership validation.")
	}
	if !ok {
		return nil, http.StatusBadRequest, errors.Wrap(err, "The organization does not own this workflow.")
	}

	return &watchWorkflowArgs{
		CommonArgs: common,
		workflowId: workflowId,
	}, http.StatusOK, nil
}

func (h *WatchWorkflowHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*watchWorkflowArgs)

	response := utils.EmptyResponse{}

	_, err := h.WorkflowWatcherWriter.CreateWorkflowWatcher(ctx, args.workflowId, args.Id, h.Database)
	if err != nil {
		return response, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error while updating the database.")
	}

	return response, http.StatusOK, nil
}
