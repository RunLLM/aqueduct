package utils

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/github"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

// WriteDAGToDatabase writes dag to the database along with all of its
// Operators, Artifacts, and DAGEdges. If the Workflow that dag is associated
// with does not exists, it also creates a new Workflow.
// It MODIFIES dag by initializing dag.ID and dag.WorkflowID (if a new Workflow)
// was created. It return the ID of the Workflow associated with dag.
func WriteDAGToDatabase(
	ctx context.Context,
	dag *models.DAG,
	workflowRepo repos.Workflow,
	dagRepo repos.DAG,
	operatorRepo repos.Operator,
	dagEdgeRepo repos.DAGEdge,
	artifactRepo repos.Artifact,
	DB database.Database,
) (uuid.UUID, error) {
	exists, err := workflowRepo.Exists(ctx, dag.WorkflowID, DB)
	if err != nil {
		return uuid.Nil, errors.Wrap(err, "Unable to check if the workflow already exists.")
	}

	if !exists {
		workflow, err := workflowRepo.Create(
			ctx,
			dag.Metadata.UserID,
			dag.Metadata.Name,
			dag.Metadata.Description,
			&dag.Metadata.Schedule,
			&dag.Metadata.RetentionPolicy,
			dag.Metadata.NotificationSettings,
			DB,
		)
		if err != nil {
			return uuid.Nil, errors.Wrap(err, "Unable to create workflow in the database.")
		}

		// Sets WorkflowID
		dag.WorkflowID = workflow.ID
	}

	newDAG, err := dagRepo.Create(
		ctx,
		dag.WorkflowID,
		&dag.StorageConfig,
		&dag.EngineConfig,
		DB,
	)
	if err != nil {
		return uuid.Nil, errors.Wrap(err, "Unable to create workflow dag in the database.")
	}

	// Sets ID
	dag.ID = newDAG.ID

	localArtifactIdToDbArtifactId := make(map[uuid.UUID]uuid.UUID, len(dag.Artifacts))

	for id, artifact := range dag.Artifacts {
		exists, err := artifactRepo.Exists(ctx, id, DB)
		if err != nil {
			return uuid.Nil, errors.Wrap(err, "Unable to check if artifact exists in database.")
		}

		dbArtifactId := id
		if !exists {
			dbArtifact, err := artifactRepo.Create(
				ctx,
				artifact.Name,
				artifact.Description,
				artifact.Type,
				DB,
			)
			if err != nil {
				return uuid.Nil, errors.Wrap(err, "Unable to create artifact in the database.")
			}

			dbArtifactId = dbArtifact.ID
		}

		localArtifactIdToDbArtifactId[artifact.ID] = dbArtifactId
	}

	for id, operator := range dag.Operators {
		exists, err := operatorRepo.Exists(ctx, id, DB)
		if err != nil {
			return uuid.Nil, errors.Wrap(err, "Unable to check if operator exists in database.")
		}

		dbOperatorId := id
		var envId *uuid.UUID = nil
		if !operator.ExecutionEnvironmentID.IsNull {
			envId = &operator.ExecutionEnvironmentID.UUID
		}

		if !exists {
			dbOperator, err := operatorRepo.Create(
				ctx,
				operator.Name,
				operator.Description,
				&operator.Spec,
				envId,
				DB,
			)
			if err != nil {
				return uuid.Nil, errors.Wrap(err, "Unable to create operator in the database.")
			}

			dbOperatorId = dbOperator.ID
		}

		for i, artifactId := range operator.Inputs {
			_, err = dagEdgeRepo.Create(
				ctx,
				newDAG.ID,
				shared.ArtifactToOperatorDAGEdge,
				localArtifactIdToDbArtifactId[artifactId],
				dbOperatorId,
				int16(i), // idx
				DB,
			)
			if err != nil {
				return uuid.Nil, errors.Wrap(err, "Unable to create workflow dag edge in the database.")
			}
		}

		for i, artifactId := range operator.Outputs {
			_, err = dagEdgeRepo.Create(
				ctx,
				newDAG.ID,
				shared.OperatorToArtifactDAGEdge,
				dbOperatorId,
				localArtifactIdToDbArtifactId[artifactId],
				int16(i), // idx
				DB,
			)
			if err != nil {
				return uuid.Nil, errors.Wrap(err, "Unable to create workflow dag edge in the database.")
			}
		}
	}

	return dag.WorkflowID, nil
}

// ReadDAGFromDatabase returns the specified DAG after initializing all
// of its persistent AND in-memory fields by making the appropriate database reads.
func ReadDAGFromDatabase(
	ctx context.Context,
	dagID uuid.UUID,
	workflowRepo repos.Workflow,
	dagRepo repos.DAG,
	operatorRepo repos.Operator,
	artifactRepo repos.Artifact,
	dagEdgeRepo repos.DAGEdge,
	DB database.Database,
) (*models.DAG, error) {
	dag, err := dagRepo.Get(ctx, dagID, DB)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read workflow dag from the database.")
	}

	workflow, err := workflowRepo.Get(ctx, dag.WorkflowID, DB)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read workflow from the database.")
	}

	dag.Metadata = workflow

	dag.Operators = make(map[uuid.UUID]models.Operator)
	dag.Artifacts = make(map[uuid.UUID]models.Artifact)

	// Populate nodes for operators and artifacts.
	operators, err := operatorRepo.GetByDAG(ctx, dag.ID, DB)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read operators from the database.")
	}

	for _, op := range operators {
		// The 'Pydantic' library on the SDK expects to receive empty lists instead of nil.
		if op.Inputs == nil {
			op.Inputs = []uuid.UUID{}
		}
		if op.Outputs == nil {
			op.Outputs = []uuid.UUID{}
		}
		dag.Operators[op.ID] = op
	}

	artifacts, err := artifactRepo.GetByDAG(ctx, dag.ID, DB)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read artifacts from the database.")
	}

	for _, artifact := range artifacts {
		dag.Artifacts[artifact.ID] = artifact
	}

	// Populate edges for operators and artifacts.
	operatorToArtifactEdges, err := dagEdgeRepo.GetOperatorToArtifactByDAG(ctx, dag.ID, DB)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read operator to artifact edges from the database.")
	}

	for _, edge := range operatorToArtifactEdges {
		if operator, ok := dag.Operators[edge.FromID]; ok {
			operator.Outputs = append(operator.Outputs, edge.ToID)
			dag.Operators[edge.FromID] = operator
		} else {
			return nil, errors.Wrap(err, "Found a dag edge with an orphaned operator id.")
		}
	}

	artifactToOperatorEdges, err := dagEdgeRepo.GetArtifactToOperatorByDAG(ctx, dag.ID, DB)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read artifact to operator edges from the database.")
	}

	for _, edge := range artifactToOperatorEdges {
		if operator, ok := dag.Operators[edge.ToID]; ok {
			operator.Inputs = append(operator.Inputs, edge.FromID)
			dag.Operators[edge.ToID] = operator
		} else {
			return nil, errors.Wrap(err, "Found a dag edge with an orphaned operator id.")
		}
	}

	return dag, nil
}

// ReadLatestDAGFromDatabase returns the latest DAG of the specified Workflow
// after initializing all of its persistent AND in-memory fields by making the
// appropriate database reads.
func ReadLatestDAGFromDatabase(
	ctx context.Context,
	workflowID uuid.UUID,
	workflowRepo repos.Workflow,
	dagRepo repos.DAG,
	operatorRepo repos.Operator,
	artifactRepo repos.Artifact,
	dagEdgeRepo repos.DAGEdge,
	DB database.Database,
) (*models.DAG, error) {
	dag, err := dagRepo.GetLatestByWorkflow(ctx, workflowID, DB)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read the latest workflow dag from the database.")
	}

	return ReadDAGFromDatabase(
		ctx,
		dag.ID,
		workflowRepo,
		dagRepo,
		operatorRepo,
		artifactRepo,
		dagEdgeRepo,
		DB,
	)
}

// This function runs 'background' update of the given workflow dag, to construct the latest version.
// For now, we only examine all github related operators and make sure we are using the latest commits.
// Any operator with newer github commits will be updated.
//
// This function updates the `workflowDag` object in-place, together with the data model updates.
// In other words, it returns the original UUID if no update happens, or the updated UUID if any part of the dag is updated.
func UpdateWorkflowDagToLatest(
	ctx context.Context,
	githubClient github.Client,
	dag *models.DAG,
	workflowRepo repos.Workflow,
	dagRepo repos.DAG,
	operatorRepo repos.Operator,
	dagEdgeRepo repos.DAGEdge,
	artifactRepo repos.Artifact,
	DB database.Database,
) (*models.DAG, error) {
	operatorsToReplace := make([]models.Operator, 0, len(dag.Operators))
	for _, op := range dag.Operators {
		opUpdated, err := github.PullOperator(
			ctx,
			githubClient,
			&op.Spec,
			&dag.StorageConfig,
		)
		if err != nil {
			return nil, err
		}

		if opUpdated {
			operatorsToReplace = append(operatorsToReplace, op)
		}
	}

	// Not updated
	if len(operatorsToReplace) == 0 {
		return dag, nil
	}

	// Update workflowDag object together with the data model.
	for _, op := range operatorsToReplace {
		delete(dag.Operators, op.ID)
		op.ID = uuid.New()
		dag.Operators[op.ID] = op
	}

	workflowID, err := WriteDAGToDatabase(
		ctx,
		dag,
		workflowRepo,
		dagRepo,
		operatorRepo,
		dagEdgeRepo,
		artifactRepo,
		DB,
	)
	if err != nil {
		return nil, err
	}

	return ReadLatestDAGFromDatabase(
		ctx,
		workflowID,
		workflowRepo,
		dagRepo,
		operatorRepo,
		artifactRepo,
		dagEdgeRepo,
		DB,
	)
}

// UpdateDAGResultMetadata updates the status and execution state of the
// specified DAGResult. It also creates the relevant notification(s).
func UpdateDAGResultMetadata(
	ctx context.Context,
	dagResultID uuid.UUID,
	execState *shared.ExecutionState,
	dagResultRepo repos.DAGResult,
	workflowRepo repos.Workflow,
	notificationRepo repos.Notification,
	DB database.Database,
) error {
	txn, err := DB.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer database.TxnRollbackIgnoreErr(ctx, txn)

	changes := map[string]interface{}{
		models.DAGResultStatus:    execState.Status,
		models.DAGResultExecState: execState,
	}

	dagResult, err := dagResultRepo.Update(
		ctx,
		dagResultID,
		changes,
		txn,
	)
	if err != nil {
		return err
	}

	if err := createDAGResultNotification(
		ctx,
		dagResult,
		notificationRepo,
		workflowRepo,
		txn,
	); err != nil {
		return err
	}

	return txn.Commit(ctx)
}

func createDAGResultNotification(
	ctx context.Context,
	dagResult *models.DAGResult,
	notificationRepo repos.Notification,
	workflowRepo repos.Workflow,
	DB database.Database,
) error {
	status := dagResult.Status
	if status != shared.SucceededExecutionStatus &&
		status != shared.FailedExecutionStatus {
		// Do not create notifications for DAGResults still in progress
		return nil
	}

	workflow, err := workflowRepo.GetByDAG(
		ctx,
		dagResult.DagID,
		DB,
	)
	if err != nil {
		return err
	}

	notificationLevel := shared.SuccessNotificationLevel
	notificationContent := fmt.Sprintf(
		"Workflow %s has succeeded!",
		workflow.Name,
	)
	if status == shared.FailedExecutionStatus {
		notificationLevel = shared.ErrorNotificationLevel
		notificationContent = fmt.Sprintf(
			"Workflow %s has failed.",
			workflow.Name,
		)
	}

	notificationAssociation := &shared.NotificationAssociation{
		Object: shared.DAGResultNotificationObject,
		ID:     dagResult.ID,
	}

	// TODO: Create notification for all watchers
	// Right now there is only 1 User in the system for only 1 notification
	// needs to be created
	_, err = notificationRepo.Create(
		ctx,
		workflow.UserID,
		notificationContent,
		notificationLevel,
		notificationAssociation,
		DB,
	)
	return err
}
