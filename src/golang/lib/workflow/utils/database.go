package utils

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/collections/notification"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_edge"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	mdl_shared "github.com/aqueducthq/aqueduct/lib/models/shared"
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
	operatorReader operator.Reader,
	operatorWriter operator.Writer,
	workflowDagEdgeWriter workflow_dag_edge.Writer,
	artifactReader artifact.Reader,
	artifactWriter artifact.Writer,
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
		exists, err := artifactReader.Exists(ctx, id, DB)
		if err != nil {
			return uuid.Nil, errors.Wrap(err, "Unable to check if artifact exists in database.")
		}

		dbArtifactId := id
		if !exists {
			dbArtifact, err := artifactWriter.CreateArtifact(
				ctx,
				artifact.Name,
				artifact.Description,
				artifact.Type,
				DB,
			)
			if err != nil {
				return uuid.Nil, errors.Wrap(err, "Unable to create artifact in the database.")
			}

			dbArtifactId = dbArtifact.Id
		}

		localArtifactIdToDbArtifactId[artifact.Id] = dbArtifactId
	}

	for id, operator := range dag.Operators {
		exists, err := operatorReader.Exists(ctx, id, DB)
		if err != nil {
			return uuid.Nil, errors.Wrap(err, "Unable to check if operator exists in database.")
		}

		dbOperatorId := id
		var envId *uuid.UUID = nil
		if !operator.ExecutionEnvironmentID.IsNull {
			envId = &operator.ExecutionEnvironmentID.UUID
		}

		if !exists {
			dbOperator, err := operatorWriter.CreateOperator(
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

			dbOperatorId = dbOperator.Id
		}

		for i, artifactId := range operator.Inputs {
			_, err = workflowDagEdgeWriter.CreateWorkflowDagEdge(
				ctx,
				newDAG.ID,
				workflow_dag_edge.ArtifactToOperatorType,
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
			_, err = workflowDagEdgeWriter.CreateWorkflowDagEdge(
				ctx,
				newDAG.ID,
				workflow_dag_edge.OperatorToArtifactType,
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
	operatorReader operator.Reader,
	artifactReader artifact.Reader,
	workflowDagEdgeReader workflow_dag_edge.Reader,
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

	dag.Operators = make(map[uuid.UUID]operator.DBOperator)
	dag.Artifacts = make(map[uuid.UUID]artifact.DBArtifact)

	// Populate nodes for operators and artifacts.
	operators, err := operatorReader.GetOperatorsByWorkflowDagId(ctx, dag.ID, DB)
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
		dag.Operators[op.Id] = op
	}

	artifacts, err := artifactReader.GetArtifactsByWorkflowDagId(ctx, dag.ID, DB)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read artifacts from the database.")
	}

	for _, artifact := range artifacts {
		dag.Artifacts[artifact.Id] = artifact
	}

	// Populate edges for operators and artifacts.
	operatorToArtifactEdges, err := workflowDagEdgeReader.GetOperatorToArtifactEdges(ctx, dag.ID, DB)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read operator to artifact edges from the database.")
	}

	for _, edge := range operatorToArtifactEdges {
		if operator, ok := dag.Operators[edge.FromId]; ok {
			operator.Outputs = append(operator.Outputs, edge.ToId)
			dag.Operators[edge.FromId] = operator
		} else {
			return nil, errors.Wrap(err, "Found a dag edge with an orphaned operator id.")
		}
	}

	artifactToOperatorEdges, err := workflowDagEdgeReader.GetArtifactToOperatorEdges(ctx, dag.ID, DB)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read artifact to operator edges from the database.")
	}

	for _, edge := range artifactToOperatorEdges {
		if operator, ok := dag.Operators[edge.ToId]; ok {
			operator.Inputs = append(operator.Inputs, edge.FromId)
			dag.Operators[edge.ToId] = operator
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
	operatorReader operator.Reader,
	artifactReader artifact.Reader,
	workflowDagEdgeReader workflow_dag_edge.Reader,
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
		operatorReader,
		artifactReader,
		workflowDagEdgeReader,
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
	operatorReader operator.Reader,
	operatorWriter operator.Writer,
	workflowDagEdgeReader workflow_dag_edge.Reader,
	workflowDagEdgeWriter workflow_dag_edge.Writer,
	artifactReader artifact.Reader,
	artifactWriter artifact.Writer,
	DB database.Database,
) (*models.DAG, error) {
	operatorsToReplace := make([]operator.DBOperator, 0, len(dag.Operators))
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
		delete(dag.Operators, op.Id)
		op.Id = uuid.New()
		dag.Operators[op.Id] = op
	}

	workflowID, err := WriteDAGToDatabase(
		ctx,
		dag,
		workflowRepo,
		dagRepo,
		operatorReader,
		operatorWriter,
		workflowDagEdgeWriter,
		artifactReader,
		artifactWriter,
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
		operatorReader,
		artifactReader,
		workflowDagEdgeReader,
		DB,
	)
}

// UpdateDAGResultMetadata updates the status and execution state of the
// specified DAGResult. It also creates the relevant notification(s).
func UpdateDAGResultMetadata(
	ctx context.Context,
	dagResultID uuid.UUID,
	execState *mdl_shared.ExecutionState,
	dagResultRepo repos.DAGResult,
	workflowRepo repos.Workflow,
	notificationWriter notification.Writer,
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
		notificationWriter,
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
	notificationWriter notification.Writer,
	workflowRepo repos.Workflow,
	DB database.Database,
) error {
	status := dagResult.Status
	if status != mdl_shared.SucceededExecutionStatus &&
		status != mdl_shared.FailedExecutionStatus {
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

	notificationLevel := notification.SuccessLevel
	notificationContent := fmt.Sprintf(
		"Workflow %s has succeeded!",
		workflow.Name,
	)
	if status == mdl_shared.FailedExecutionStatus {
		notificationLevel = notification.ErrorLevel
		notificationContent = fmt.Sprintf(
			"Workflow %s has failed.",
			workflow.Name,
		)
	}

	notificationAssociation := notification.NotificationAssociation{
		Object: notification.WorkflowDagResultObject,
		Id:     dagResult.ID,
	}

	// TODO: Create notification for all watchers
	// Right now there is only 1 User in the system for only 1 notification
	// needs to be created
	_, err = notificationWriter.CreateNotification(
		ctx,
		workflow.UserID,
		notificationContent,
		notificationLevel,
		notificationAssociation,
		DB,
	)
	return err
}
