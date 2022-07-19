package dag

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/collections/artifact_result"
	"github.com/aqueducthq/aqueduct/lib/collections/notification"
	"github.com/aqueducthq/aqueduct/lib/collections/operator_result"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/collections/user"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_result"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/artifact"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator"
	"github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

type WorkflowDag interface {
	Operators() map[uuid.UUID]operator.Operator
	Artifacts() map[uuid.UUID]artifact.Artifact

	// OperatorsOnArtifact returns all the operators that consume the given artifact as input.
	OperatorsOnArtifact(artifact.Artifact) ([]operator.Operator, error)

	// ArtifactsFromOperator returns all the artifacts produced by the given operator.
	OperatorOutputs(operator.Operator) ([]artifact.Artifact, error)

	// OperatorInputs returns all the artifacts that are fed as input to the given operator.
	OperatorInputs(operator.Operator) ([]artifact.Artifact, error)

	// InitializeResult initializes the dag result in the database.
	// Also initializes the operators and artifacts contained in this dag.
	InitializeResults(ctx context.Context) error

	// PersistResult updates the dag result in the database after execution.
	// InitializeResult() must have already been called.
	// *Does not* persist the operators or artifacts contained in this dag.
	PersistResult(ctx context.Context, status shared.ExecutionStatus) error
}

type workflowDagImpl struct {
	dbWorkflowDag *workflow_dag.DBWorkflowDag

	operators           map[uuid.UUID]operator.Operator
	artifacts           map[uuid.UUID]artifact.Artifact
	opToOutputArtifacts map[uuid.UUID][]uuid.UUID
	opToInputArtifacts  map[uuid.UUID][]uuid.UUID
	artifactToOps       map[uuid.UUID][]uuid.UUID

	resultWriter       workflow_dag_result.Writer
	workflowReader     workflow.Reader
	notificationWriter notification.Writer
	userReader         user.Reader
	db                 database.Database

	// Corresponds to the workflow dag result entry in the database.
	// This is empty if InitializeResultsj() has not been called.
	resultID uuid.UUID
}

func NewWorkflowDag(
	ctx context.Context,
	dbWorkflowDag *workflow_dag.DBWorkflowDag,
	dagResultWriter workflow_dag_result.Writer,
	opResultWriter operator_result.Writer,
	artifactResultWriter artifact_result.Writer,
	workflowReader workflow.Reader,
	notificationWriter notification.Writer,
	userReader user.Reader,
	jobManager job.JobManager,
	vaultObject vault.Vault,
	storageConfig *shared.StorageConfig,
	db database.Database,
) (WorkflowDag, error) {
	// First, allocate a content and metadata path for each artifact.
	artifactIDToContentPath := make(map[uuid.UUID]string, len(dbWorkflowDag.Artifacts))
	artifactIDToMetadataPath := make(map[uuid.UUID]string, len(dbWorkflowDag.Artifacts))
	for _, dbArtifact := range dbWorkflowDag.Artifacts {
		artifactIDToContentPath[dbArtifact.Id] = uuid.New().String()
		artifactIDToMetadataPath[dbArtifact.Id] = uuid.New().String()
	}

	var err error

	// With all the initial database writes completed (if at all), we can now initialize
	// the operator and artifact classes. As well as the connections between them.
	operators := make(map[uuid.UUID]operator.Operator, len(dbWorkflowDag.Operators))
	artifacts := make(map[uuid.UUID]artifact.Artifact, len(dbWorkflowDag.Artifacts))
	for artifactID, dbArtifact := range dbWorkflowDag.Artifacts {
		artifacts[artifactID], err = artifact.NewArtifact(
			dbArtifact,
			artifactIDToContentPath[artifactID],
			artifactIDToMetadataPath[artifactID],
			artifactResultWriter,
			storageConfig,
			db,
		)
		if err != nil {
			return nil, err
		}
	}

	// These artifact <-> operator maps help us remember all dag connections.
	artifactToOps := make(map[uuid.UUID][]uuid.UUID, len(artifacts))
	for artifactID := range artifacts {
		artifactToOps[artifactID] = make([]uuid.UUID, 0, 1)
	}
	opToOutputArtifacts := make(map[uuid.UUID][]uuid.UUID, len(operators))
	opToInputArtifacts := make(map[uuid.UUID][]uuid.UUID, len(operators))
	for opID, dbOperator := range dbWorkflowDag.Operators {
		opToOutputArtifacts[opID] = make([]uuid.UUID, 0, 1)
		opToInputArtifacts[opID] = make([]uuid.UUID, 0, 1)

		inputArtifacts := make([]artifact.Artifact, 0, len(artifacts))
		inputContentPaths := make([]string, 0, len(dbOperator.Inputs))
		inputMetadataPaths := make([]string, 0, len(dbOperator.Inputs))
		for _, artifactID := range dbOperator.Inputs {
			inputArtifacts = append(inputArtifacts, artifacts[artifactID])
			inputContentPaths = append(inputContentPaths, artifactIDToContentPath[artifactID])
			inputMetadataPaths = append(inputMetadataPaths, artifactIDToMetadataPath[artifactID])

			artifactToOps[artifactID] = append(artifactToOps[artifactID], opID)
			opToInputArtifacts[opID] = append(opToInputArtifacts[opID], artifactID)
		}
		outputArtifacts := make([]artifact.Artifact, 0, len(artifacts))
		outputContentPaths := make([]string, 0, len(dbOperator.Outputs))
		outputMetadataPaths := make([]string, 0, len(dbOperator.Outputs))
		for _, artifactID := range dbOperator.Outputs {
			outputArtifacts = append(outputArtifacts, artifacts[artifactID])
			outputContentPaths = append(outputContentPaths, artifactIDToContentPath[artifactID])
			outputMetadataPaths = append(outputMetadataPaths, artifactIDToMetadataPath[artifactID])

			opToOutputArtifacts[opID] = append(opToOutputArtifacts[opID], artifactID)
		}

		operators[opID], err = operator.NewOperator(
			ctx,
			dbOperator,
			inputArtifacts,
			inputContentPaths,
			inputMetadataPaths,
			outputArtifacts,
			outputContentPaths,
			outputMetadataPaths,
			opResultWriter,
			jobManager,
			vaultObject,
			storageConfig,
			db,
		)
		if err != nil {
			return nil, err
		}
	}

	return &workflowDagImpl{
		dbWorkflowDag:       dbWorkflowDag,
		operators:           operators,
		artifacts:           artifacts,
		opToOutputArtifacts: opToOutputArtifacts,
		opToInputArtifacts:  opToInputArtifacts,
		artifactToOps:       artifactToOps,

		resultWriter:       dagResultWriter,
		workflowReader:     workflowReader,
		notificationWriter: notificationWriter,
		userReader:         userReader,
		db:                 db,
		resultID:           uuid.Nil,
	}, nil
}

func (w *workflowDagImpl) Operators() map[uuid.UUID]operator.Operator {
	return w.operators
}

func (w *workflowDagImpl) Artifacts() map[uuid.UUID]artifact.Artifact {
	return w.artifacts
}

func (w *workflowDagImpl) OperatorsOnArtifact(artifact artifact.Artifact) ([]operator.Operator, error) {
	opIDs, ok := w.artifactToOps[artifact.ID()]
	if !ok {
		return nil, errors.Newf("Unable to find artifact %s (%s) on dag.", artifact.ID(), artifact.Name())
	}

	ops := make([]operator.Operator, 0, len(opIDs))
	for _, opID := range opIDs {
		ops = append(ops, w.operators[opID])
	}
	return ops, nil
}

func (w *workflowDagImpl) OperatorOutputs(op operator.Operator) ([]artifact.Artifact, error) {
	artifactIDs, ok := w.opToOutputArtifacts[op.ID()]
	if !ok {
		return nil, errors.Newf("Unable to find operator %s (%s) on dag.", op.ID(), op.Name())
	}

	artifacts := make([]artifact.Artifact, 0, len(artifactIDs))
	for _, artifactID := range artifactIDs {
		artifacts = append(artifacts, w.artifacts[artifactID])
	}
	return artifacts, nil
}

func (w *workflowDagImpl) OperatorInputs(op operator.Operator) ([]artifact.Artifact, error) {
	artifactIDs, ok := w.opToInputArtifacts[op.ID()]
	if !ok {
		return nil, errors.Newf("Unable to find operator %s (%s) on dag.", op.ID(), op.Name())
	}

	artifacts := make([]artifact.Artifact, 0, len(artifactIDs))
	for _, artifactID := range artifactIDs {
		artifacts = append(artifacts, w.artifacts[artifactID])
	}
	return artifacts, nil
}

func (w *workflowDagImpl) InitializeResults(ctx context.Context) error {
	if w.resultWriter == nil {
		return errors.New("Workflow dag's result writer cannot be nil.")
	}

	// Create a database record of workflow dag result and set its status to `pending`.
	// TODO(ENG-599): wrap these writes into a transaction.
	dagResult, err := w.resultWriter.CreateWorkflowDagResult(ctx, w.dbWorkflowDag.Id, w.db)
	if err != nil {
		return errors.Wrap(err, "Unable to create workflow dag result record.")
	}
	w.resultID = dagResult.Id

	// Also initialize the operators and artifact results.
	for _, op := range w.Operators() {
		err := op.InitializeResult(ctx, w.resultID)
		if err != nil {
			return err
		}
	}
	for _, artf := range w.Artifacts() {
		err := artf.InitializeResult(ctx, w.resultID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *workflowDagImpl) PersistResult(ctx context.Context, status shared.ExecutionStatus) error {
	if w.resultID == uuid.Nil {
		return errors.New("Workflow's dag result was not initialized before calling PersistResult.")
	}

	// We `defer` this call to ensure that the WorkflowDagResult metadata is always updated.
	utils.UpdateWorkflowDagResultMetadata(
		ctx,
		w.resultID,
		status,
		w.resultWriter,
		w.workflowReader,
		w.notificationWriter,
		w.userReader,
		w.db,
	)

	return nil
}
