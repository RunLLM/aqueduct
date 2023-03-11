package workflow

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/aqueducthq/aqueduct/lib/graph"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/google/uuid"
)

const (
	internalValidationErrMsg = "Internal system error occurred while validating Workflow schedule."
)

// ValidateSchedule validates the provided schedule. The flag isUpdate indicates
// whether or not the Workflow, for which the schedule is being validated already exists.
// If the Workflow exists, its ID should be provided as workflowID; otherwise,
// workflowID can be ignored.
// The following conditions are not allowed:
// 1. Having a CascadingUpdateTrigger where SourceID is for a Workflow that is
// running on a non self-orchestrated engine, such as Airflow. This is not allowed
// since Aqueduct cannot trigger Workflow runs at the end of execution on an
// engine that is not self-orchestrated.
// 2. Having a CascadingUpdateTrigger that creates a cycle amongst the cascading workflows.
// It returns an HTTP status code and a client-friendly error, if any.
func ValidateSchedule(
	ctx context.Context,
	isUpdate bool,
	workflowID uuid.UUID,
	schedule shared.Schedule,
	engineType shared.EngineType,
	artifactRepo repos.Artifact,
	dagRepo repos.DAG,
	dagEdgeRepo repos.DAGEdge,
	operatorRepo repos.Operator,
	workflowRepo repos.Workflow,
	DB database.Database,
) (int, error) {
	if schedule.Trigger != shared.CascadingUpdateTrigger {
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

	sourceDAG, err := utils.ReadLatestDAGFromDatabase(
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

	// Condition 1
	if sourceDAG.EngineConfig.Type == shared.AirflowEngineType {
		return http.StatusBadRequest, errors.New("Cannot use Workflows running on Airflow for the source.")
	}

	// Condition 2
	if !isUpdate {
		// It is not possible to form a cycle when registering a NEW workflow.
		// A cycle is formed when registering a workflow, B, with source, A, if
		// and only if there already exists a path from B to A. Since workflow B
		// does not exist, such a path also does not exist. In other words,
		// we are adding a new edge in a graph from nodes A to B, but since node B
		// is new, we know that there are no outgoing edges from B, so we can
		// conclude that no cycle is formed.
		return http.StatusOK, nil
	}

	// Note that we still need to check for cycles when the schedule of an existing
	// Workflow is modified.

	cascadingWorkflows, err := workflowRepo.GetByScheduleTrigger(ctx, shared.CascadingUpdateTrigger, DB)
	if err != nil {
		return http.StatusInternalServerError, errors.Wrap(err, internalValidationErrMsg)
	}

	if hasCycle := checkForCycle(workflowID, schedule.SourceID, cascadingWorkflows); hasCycle {
		return http.StatusBadRequest, errors.New("Cannot allow cycles for cascading workflows.")
	}

	return http.StatusOK, nil
}

// checkForCycle returns true if setting workflowID's source workflow to sourceID would
// result in a cycle.
func checkForCycle(workflowID uuid.UUID, sourceID uuid.UUID, targetWorkflows []models.Workflow) bool {
	gph := graph.NewDirected()
	for _, targetWorkflow := range targetWorkflows {
		gph.AddNode(targetWorkflow.ID)
		gph.AddNode(targetWorkflow.Schedule.SourceID)

		if targetWorkflow.ID != workflowID {
			// We don't add an edge for workflowID to its current source Workflow,
			// since that will be overwritten by the new source anyways.
			gph.AddEdge(targetWorkflow.Schedule.SourceID, targetWorkflow.ID)
		}
	}

	// There is a cycle if there exists a path from workflowID to sourceID
	return gph.HasPath(workflowID, sourceID)
}
