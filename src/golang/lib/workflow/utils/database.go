package utils

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/collections/notification"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_edge"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_result"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	mdl_shared "github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/github"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

func WriteWorkflowDagToDatabase(
	ctx context.Context,
	dag *workflow_dag.DBWorkflowDag,
	workflowReader workflow.Reader,
	workflowWriter workflow.Writer,
	workflowDagWriter workflow_dag.Writer,
	operatorReader operator.Reader,
	operatorWriter operator.Writer,
	workflowDagEdgeWriter workflow_dag_edge.Writer,
	artifactReader artifact.Reader,
	artifactWriter artifact.Writer,
	db database.Database,
) (uuid.UUID, error) {
	exists, err := workflowReader.Exists(ctx, dag.WorkflowId, db)
	if err != nil {
		return uuid.Nil, errors.Wrap(err, "Unable to check if the workflow already exists.")
	}

	workflowId := dag.WorkflowId
	if !exists {
		workflow, err := workflowWriter.CreateWorkflow(
			ctx,
			dag.Metadata.UserId,
			dag.Metadata.Name,
			dag.Metadata.Description,
			&dag.Metadata.Schedule,
			&dag.Metadata.RetentionPolicy,
			db,
		)
		if err != nil {
			return uuid.Nil, errors.Wrap(err, "Unable to create workflow in the database.")
		}
		workflowId = workflow.Id
	}

	workflowDag, err := workflowDagWriter.CreateWorkflowDag(
		ctx,
		workflowId,
		&dag.StorageConfig,
		&dag.EngineConfig,
		db,
	)
	if err != nil {
		return uuid.Nil, errors.Wrap(err, "Unable to create workflow dag in the database.")
	}
	dag.Id = workflowDag.Id

	localArtifactIdToDbArtifactId := make(map[uuid.UUID]uuid.UUID, len(dag.Artifacts))

	for id, artifact := range dag.Artifacts {
		exists, err := artifactReader.Exists(ctx, id, db)
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
				db,
			)
			if err != nil {
				return uuid.Nil, errors.Wrap(err, "Unable to create artifact in the database.")
			}

			dbArtifactId = dbArtifact.Id
		}

		localArtifactIdToDbArtifactId[artifact.Id] = dbArtifactId
	}

	for id, operator := range dag.Operators {
		exists, err := operatorReader.Exists(ctx, id, db)
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
				db,
			)
			if err != nil {
				return uuid.Nil, errors.Wrap(err, "Unable to create operator in the database.")
			}

			dbOperatorId = dbOperator.Id
		}

		for i, artifactId := range operator.Inputs {
			_, err = workflowDagEdgeWriter.CreateWorkflowDagEdge(
				ctx,
				workflowDag.Id,
				workflow_dag_edge.ArtifactToOperatorType,
				localArtifactIdToDbArtifactId[artifactId],
				dbOperatorId,
				int16(i), // idx
				db,
			)
			if err != nil {
				return uuid.Nil, errors.Wrap(err, "Unable to create workflow dag edge in the database.")
			}
		}

		for i, artifactId := range operator.Outputs {
			_, err = workflowDagEdgeWriter.CreateWorkflowDagEdge(
				ctx,
				workflowDag.Id,
				workflow_dag_edge.OperatorToArtifactType,
				dbOperatorId,
				localArtifactIdToDbArtifactId[artifactId],
				int16(i), // idx
				db,
			)
			if err != nil {
				return uuid.Nil, errors.Wrap(err, "Unable to create workflow dag edge in the database.")
			}
		}
	}

	return workflowId, nil
}

func ReadWorkflowDagFromDatabase(
	ctx context.Context,
	workflowDagId uuid.UUID,
	workflowReader workflow.Reader,
	workflowDagReader workflow_dag.Reader,
	operatorReader operator.Reader,
	artifactReader artifact.Reader,
	workflowDagEdgeReader workflow_dag_edge.Reader,
	db database.Database,
) (*workflow_dag.DBWorkflowDag, error) {
	workflowDag, err := workflowDagReader.GetWorkflowDag(ctx, workflowDagId, db)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read workflow dag from the database.")
	}

	dbWorkflow, err := workflowReader.GetWorkflow(ctx, workflowDag.WorkflowId, db)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read workflow from the database.")
	}

	workflowDag.Metadata = dbWorkflow

	workflowDag.Operators = make(map[uuid.UUID]operator.DBOperator)
	workflowDag.Artifacts = make(map[uuid.UUID]artifact.DBArtifact)

	// Populate nodes for operators and artifacts.
	operators, err := operatorReader.GetOperatorsByWorkflowDagId(ctx, workflowDag.Id, db)
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
		workflowDag.Operators[op.Id] = op
	}

	artifacts, err := artifactReader.GetArtifactsByWorkflowDagId(ctx, workflowDag.Id, db)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read artifacts from the database.")
	}

	for _, artifact := range artifacts {
		workflowDag.Artifacts[artifact.Id] = artifact
	}

	// Populate edges for operators and artifacts.
	operatorToArtifactEdges, err := workflowDagEdgeReader.GetOperatorToArtifactEdges(ctx, workflowDag.Id, db)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read operator to artifact edges from the database.")
	}

	for _, edge := range operatorToArtifactEdges {
		if operator, ok := workflowDag.Operators[edge.FromId]; ok {
			operator.Outputs = append(operator.Outputs, edge.ToId)
			workflowDag.Operators[edge.FromId] = operator
		} else {
			return nil, errors.Wrap(err, "Found a dag edge with an orphaned operator id.")
		}
	}

	artifactToOperatorEdges, err := workflowDagEdgeReader.GetArtifactToOperatorEdges(ctx, workflowDag.Id, db)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read artifact to operator edges from the database.")
	}

	for _, edge := range artifactToOperatorEdges {
		if operator, ok := workflowDag.Operators[edge.ToId]; ok {
			operator.Inputs = append(operator.Inputs, edge.FromId)
			workflowDag.Operators[edge.ToId] = operator
		} else {
			return nil, errors.Wrap(err, "Found a dag edge with an orphaned operator id.")
		}
	}

	return workflowDag, nil
}

func ReadLatestWorkflowDagFromDatabase(
	ctx context.Context,
	workflowId uuid.UUID,
	workflowReader workflow.Reader,
	workflowDagReader workflow_dag.Reader,
	operatorReader operator.Reader,
	artifactReader artifact.Reader,
	workflowDagEdgeReader workflow_dag_edge.Reader,
	db database.Database,
) (*workflow_dag.DBWorkflowDag, error) {
	workflowDag, err := workflowDagReader.GetLatestWorkflowDag(ctx, workflowId, db)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read the latest workflow dag from the database.")
	}

	return ReadWorkflowDagFromDatabase(
		ctx,
		workflowDag.Id,
		workflowReader,
		workflowDagReader,
		operatorReader,
		artifactReader,
		workflowDagEdgeReader,
		db,
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
	workflowDag *workflow_dag.DBWorkflowDag,
	workflowReader workflow.Reader,
	workflowWriter workflow.Writer,
	workflowDagReader workflow_dag.Reader,
	workflowDagWriter workflow_dag.Writer,
	operatorReader operator.Reader,
	operatorWriter operator.Writer,
	workflowDagEdgeReader workflow_dag_edge.Reader,
	workflowDagEdgeWriter workflow_dag_edge.Writer,
	artifactReader artifact.Reader,
	artifactWriter artifact.Writer,
	db database.Database,
) (*workflow_dag.DBWorkflowDag, error) {
	operatorsToReplace := make([]operator.DBOperator, 0, len(workflowDag.Operators))
	for _, op := range workflowDag.Operators {
		opUpdated, err := github.PullOperator(
			ctx,
			githubClient,
			&op.Spec,
			&workflowDag.StorageConfig,
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
		return workflowDag, nil
	}

	// Update workflowDag object together with the data model.
	for _, op := range operatorsToReplace {
		delete(workflowDag.Operators, op.Id)
		op.Id = uuid.New()
		workflowDag.Operators[op.Id] = op
	}

	workflowId, err := WriteWorkflowDagToDatabase(
		ctx,
		workflowDag,
		workflowReader,
		workflowWriter,
		workflowDagWriter,
		operatorReader,
		operatorWriter,
		workflowDagEdgeWriter,
		artifactReader,
		artifactWriter,
		db,
	)
	if err != nil {
		return nil, err
	}

	return ReadLatestWorkflowDagFromDatabase(
		ctx,
		workflowId,
		workflowReader,
		workflowDagReader,
		operatorReader,
		artifactReader,
		workflowDagEdgeReader,
		db,
	)
}

func CreateWorkflowDagResult(
	ctx context.Context,
	workflowDagId uuid.UUID,
	execState *shared.ExecutionState,
	workflowDagResultWriter workflow_dag_result.Writer,
	db database.Database,
) (*workflow_dag_result.WorkflowDagResult, error) {
	return workflowDagResultWriter.CreateWorkflowDagResult(
		ctx,
		workflowDagId,
		execState,
		db,
	)
}

// UpdateDAGResultMetadata updates the status and execution state of the
// specified DAGResult. It also creates the relevant notification(s).
func UpdateDAGResultMetadata(
	ctx context.Context,
	dagResultID uuid.UUID,
	execState *shared.ExecutionState,
	dagResultRepo repos.DAGResult,
	workflowReader workflow.Reader,
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
		workflowReader,
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
	workflowReader workflow.Reader,
	DB database.Database,
) error {
	status := dagResult.Status
	if status != mdl_shared.SucceededExecutionStatus &&
		status != mdl_shared.FailedExecutionStatus {
		// Do not create notifications for DAGResults still in progress
		return nil
	}

	workflow, err := workflowReader.GetWorkflowByWorkflowDagId(
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
		workflow.UserId,
		notificationContent,
		notificationLevel,
		notificationAssociation,
		DB,
	)
	return err
}
