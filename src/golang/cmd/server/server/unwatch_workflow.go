package server

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/internal/server/utils"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_watcher"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type unwatchWorkflowArgs struct {
	*CommonArgs
	workflowId uuid.UUID
}

type UnwatchWorkflowHandler struct {
	PostHandler

	Database              database.Database
	WorkflowReader        workflow.Reader
	WorkflowWatcherWriter workflow_watcher.Writer
}

func (*UnwatchWorkflowHandler) Name() string {
	return "UnwatchWorkflow"
}

func (h *UnwatchWorkflowHandler) Prepare(r *http.Request) (interface{}, int, error) {
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

	return &unwatchWorkflowArgs{
		CommonArgs: common,
		workflowId: workflowId,
	}, http.StatusOK, nil
}

func (h *UnwatchWorkflowHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*unwatchWorkflowArgs)

	response := utils.EmptyResponse{}

	err := h.WorkflowWatcherWriter.DeleteWorkflowWatcher(ctx, args.workflowId, args.Id, h.Database)
	if err != nil {
		return response, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error while updating the database.")
	}

	return response, http.StatusOK, nil
}
