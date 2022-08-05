package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/workflow/engine"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type EditWorkflowHandler struct {
	PostHandler

	Database       database.Database
	WorkflowReader workflow.Reader
	Engine         engine.Engine
}

type editWorkflowInput struct {
	WorkflowName        string             `json:"name"`
	WorkflowDescription string             `json:"description"`
	Schedule            *workflow.Schedule `json:"schedule"`
}

type editWorkflowArgs struct {
	workflowId          uuid.UUID
	workflowName        string
	workflowDescription string
	schedule            *workflow.Schedule
}

func (*EditWorkflowHandler) Name() string {
	return "EditWorkflow"
}

func (h *EditWorkflowHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	workflowIdStr := chi.URLParam(r, routes.WorkflowIdUrlParam)
	if workflowIdStr == "" {
		return nil, http.StatusBadRequest, errors.New("No workflow id was specified.")
	}

	workflowId, err := uuid.Parse(workflowIdStr)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Malformed workflow ID.")
	}

	ok, err := h.WorkflowReader.ValidateWorkflowOwnership(
		r.Context(),
		workflowId,
		aqContext.OrganizationId,
		h.Database,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unexpected error during workflow ownership validation.")
	}
	if !ok {
		return nil, http.StatusBadRequest, errors.Wrap(err, "The organization does not own this workflow.")
	}

	var input editWorkflowInput
	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		return nil, http.StatusBadRequest, errors.New("Unable to parse JSON input.")
	}

	// First, we check if the workflow type is set to periodic. If it is, we
	// enforce that a cron schedule must be present on the schedule object,
	// otherwise we fail out. Critically, this is true whether the workflow is
	// paused or not. This is important because when we load the schedule for a
	// paused workflow, unpausing it should resume previous behavior.
	if input.Schedule.Trigger == workflow.PeriodicUpdateTrigger && input.Schedule.CronSchedule == "" {
		return nil, http.StatusBadRequest, errors.New("Invalid workflow schedule specified.")
	}

	// If the workflow is paused, it must be in periodic update mode.
	if input.Schedule.Trigger == workflow.ManualUpdateTrigger && input.Schedule.Paused {
		return nil, http.StatusBadRequest, errors.New("Cannot pause a manually updated workflow.")
	}

	// Finally, we check if there are an updates at all.
	if input.WorkflowName == "" && input.WorkflowDescription == "" && input.Schedule.Trigger == "" {
		return nil, http.StatusBadRequest, errors.New("Edit request issued without any updates specified.")
	}

	return &editWorkflowArgs{
		workflowId:          workflowId,
		workflowName:        input.WorkflowName,
		workflowDescription: input.WorkflowDescription,
		schedule:            input.Schedule,
	}, http.StatusOK, nil
}

func (h *EditWorkflowHandler) Perform(ctx context.Context, interfaceArgs interface{}) (interface{}, int, error) {
	args := interfaceArgs.(*editWorkflowArgs)
	txn, err := h.Database.BeginTx(ctx)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to update workflow.")
	}
	defer database.TxnRollbackIgnoreErr(ctx, txn)

	err = h.Engine.EditWorkflow(
		ctx,
		txn,
		args.workflowId,
		args.workflowName,
		args.workflowDescription,
		args.schedule,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to update workflow.")
	}

	if err := txn.Commit(ctx); err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to update workflow.")
	}

	return struct{}{}, http.StatusOK, nil
}
