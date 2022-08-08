package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/job"
	shared_utils "github.com/aqueducthq/aqueduct/lib/lib_utils"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/github"
	"github.com/dropbox/godropbox/errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type EditWorkflowHandler struct {
	PostHandler

	Database       database.Database
	WorkflowReader workflow.Reader
	WorkflowWriter workflow.Writer
	JobManager     job.JobManager
	Vault          vault.Vault
	GithubManager  github.Manager
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

	changes := map[string]interface{}{}
	if args.workflowName != "" {
		changes["name"] = args.workflowName
	}

	if args.workflowDescription != "" {
		changes["description"] = args.workflowDescription
	}

	if args.schedule.Trigger != "" {
		cronjobName := shared_utils.AppendPrefix(args.workflowId.String())
		err := h.updateWorkflowSchedule(ctx, args.workflowId.String(), cronjobName, args.schedule)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		changes["schedule"] = args.schedule
	}

	_, err := h.WorkflowWriter.UpdateWorkflow(ctx, args.workflowId, changes, h.Database)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "Unable to update workflow.")
	}

	return struct{}{}, http.StatusOK, nil
}

func (h *EditWorkflowHandler) updateWorkflowSchedule(
	ctx context.Context,
	workflowId string,
	cronjobName string,
	newSchedule *workflow.Schedule,
) error {
	// How we update the workflow schedule depends on whether a cron job already exists.
	// A manually triggered workflow does not have a cron job. If we're editing it to have a periodic
	// schedule, we'll need to create a new cron job.
	if !h.JobManager.CronJobExists(ctx, cronjobName) {
		if newSchedule.CronSchedule != "" {
			spec := job.NewWorkflowSpec(
				cronjobName,
				workflowId,
				h.Database.Config(),
				h.Vault.Config(),
				h.JobManager.Config(),
				h.GithubManager.Config(),
				nil, /* parameters */
			)
			err := h.JobManager.DeployCronJob(
				ctx,
				cronjobName,
				string(newSchedule.CronSchedule),
				spec,
			)
			if err != nil {
				return errors.Wrap(err, "Unable to deploy new cron job.")
			}
		}
		// We will no-op if the workflow continues to be manually triggered.
	} else {
		// Here, we can blindly set the cron job to be paused without any other
		// modification. The pausedness of the workflow will be written to the
		// database by the changes map above, and `prepare` guarantees us that
		// if `Paused` is true, then the workflow type is `Periodic`, which in
		// turn means a schedule must be set.
		newCronSchedule := string(newSchedule.CronSchedule)
		if newSchedule.Paused {
			// The `EditCronJob` helper automatically pauses a workflow when
			// you set the cron job schedule to an empty string.
			newCronSchedule = ""
		}

		err := h.JobManager.EditCronJob(ctx, cronjobName, newCronSchedule)
		if err != nil {
			return errors.Wrap(err, "Unable to change workflow schedule.")
		}
	}
	return nil
}
