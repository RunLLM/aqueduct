package workflow

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/dropbox/godropbox/errors"
)

const (
	internalValidationErrMsg = "Internal system error occurred while validating Workflow schedule."
)

// ValidateSchedule validates the provided schedule. The following conditions are
// not allowed:
// - Having a CascadingUpdateTrigger where SourceID is for a Workflow that is
// running on a non self-orchestrated engine, such as Airflow. This is not allowed
// since Aqueduct cannot trigger Workflow runs at the end of execution on an
// engine that is not self-orchestrated.
// It returns an HTTP status code and a client-friendly error, if any.
func ValidateSchedule(
	ctx context.Context,
	schedule workflow.Schedule,
	artifactRepo repos.Artifact,
	dagRepo repos.DAG,
	dagEdgeRepo repos.DAGEdge,
	operatorRepo repos.Operator,
	workflowRepo repos.Workflow,
	DB database.Database,
) (int, error) {
	if schedule.Trigger != workflow.CascadingUpdateTrigger {
		// Only CascadingUpdateTriggers require validation
		return http.StatusOK, nil
	}

	exists, err := workflowRepo.Exists(ctx, schedule.SourceID, DB)
	if err != nil {
		return http.StatusInternalServerError, errors.Wrap(err, internalValidationErrMsg)
	}

	if !exists {
		return http.StatusBadRequest, errors.New("The specified source Workflow does not exist.")
	}

	dag, err := utils.ReadLatestDAGFromDatabase(
		ctx,
		schedule.SourceID,
		workflowRepo,
		dagRepo,
		operatorRepo,
		artifactRepo,
		dagEdgeRepo,
		DB,
	)
	if err != nil {
		return http.StatusInternalServerError, errors.Wrap(err, internalValidationErrMsg)
	}

	if dag.EngineConfig.Type == shared.AirflowEngineType {
		return http.StatusBadRequest, errors.New("Cannot use Workflows running on Airflow for the source.")
	}

	return http.StatusOK, nil
}
