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

	OperatorsOnArtifact(artifact.Artifact) ([]operator.Operator, error)
	ArtifactsFromOperator(operator.Operator) ([]artifact.Artifact, error)

	PersistResult(ctx context.Context, status shared.ExecutionStatus) error
	Finish(ctx context.Context)
}

type workflowDagImpl struct {
	dbWorkflowDag *workflow_dag.DBWorkflowDag

	operators     map[uuid.UUID]operator.Operator
	artifacts     map[uuid.UUID]artifact.Artifact
	opToArtifacts map[uuid.UUID][]uuid.UUID
	artifactToOps map[uuid.UUID][]uuid.UUID

	workflowDagResultWriter workflow_dag_result.Writer
	workflowReader          workflow.Reader
	notificationWriter      notification.Writer
	userReader              user.Reader
	db                      database.Database

	// Corresponds to the workflow dag result entry in the database. This is set during construction
	// and indicates whether the workflow dag can be persisted.
	// Persist() will no-op if this is empty.
	workflowDagResultID uuid.UUID
}

func initializeDagResultInDatabase(
	ctx context.Context,
	dbWorkflowDag *workflow_dag.DBWorkflowDag,
	dagResultWriter workflow_dag_result.Writer,
	db database.Database,
) (uuid.UUID, error) {
	// Create a database record of workflow dag result and set its status to `pending`.
	// TODO(ENG-599): wrap these writes into a transaction.
	workflowDagResult, err := dagResultWriter.CreateWorkflowDagResult(ctx, dbWorkflowDag.Id, db)
	if err != nil {
		return uuid.Nil, errors.Wrap(err, "Unable to create workflow dag result record.")
	}
	return workflowDagResult.Id, nil
}

func NewWorkflowDagNoPersist(
	ctx context.Context,
	dbWorkflowDag *workflow_dag.DBWorkflowDag,
	jobManager job.JobManager,
	vaultObject vault.Vault,
	storageConfig *shared.StorageConfig,
	db database.Database,
) (WorkflowDag, error) {
	return NewWorkflowDag(
		ctx,
		dbWorkflowDag,
		workflow_dag_result.NewNoopWriter(true),
		operator_result.NewNoopWriter(true),
		artifact_result.NewNoopWriter(true),
		workflow.NewNoopReader(true),
		notification.NewNoopWriter(true),
		user.NewNoopReader(true),
		jobManager,
		vaultObject,
		storageConfig,
		db,
		false, /* canPersist */
	)
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
	canPersist bool,
) (WorkflowDag, error) {

	// First, allocate a content and metadata path for each artifact.
	artifactIDToContentPath := make(map[uuid.UUID]string, len(dbWorkflowDag.Artifacts))
	artifactIDToMetadataPath := make(map[uuid.UUID]string, len(dbWorkflowDag.Artifacts))
	for _, dbArtifact := range dbWorkflowDag.Artifacts {
		artifactIDToContentPath[dbArtifact.Id] = uuid.New().String()
		artifactIDToMetadataPath[dbArtifact.Id] = uuid.New().String()
	}

	var workflowDagResultID uuid.UUID
	var err error
	if canPersist {
		if dagResultWriter == nil || opResultWriter == nil || artifactResultWriter == nil {
			return nil, errors.New("Nil was supplied for a database writer.")
		}

		workflowDagResultID, err = initializeDagResultInDatabase(
			ctx,
			dbWorkflowDag,
			dagResultWriter,
			db,
		)
		if err != nil {
			return nil, err
		}
	}

	// With all the initial database writes completed (if at all), we can now initialize
	// the operator and artifact classes. As well as the connections between them.
	operators := make(map[uuid.UUID]operator.Operator, len(dbWorkflowDag.Operators))
	artifacts := make(map[uuid.UUID]artifact.Artifact, len(dbWorkflowDag.Artifacts))
	for artifactID, dbArtifact := range dbWorkflowDag.Artifacts {
		artifacts[artifactID], err = artifact.NewArtifact(
			ctx,
			dbArtifact,
			artifactIDToContentPath[artifactID],
			artifactIDToMetadataPath[artifactID],
			artifactResultWriter,
			workflowDagResultID,
			db,
		)
		if err != nil {
			return nil, err
		}
	}

	// These two maps allow us to remember all the dag connections.
	artifactToOps := make(map[uuid.UUID][]uuid.UUID, len(artifacts))
	for artifactID, _ := range artifacts {
		artifactToOps[artifactID] = make([]uuid.UUID, 0, 1)
	}
	opToArtifacts := make(map[uuid.UUID][]uuid.UUID, len(operators))
	for opID, _ := range operators {
		opToArtifacts[opID] = make([]uuid.UUID, 0, 1)
	}

	for opID, dbOperator := range dbWorkflowDag.Operators {
		inputArtifacts := make([]artifact.Artifact, 0, len(artifacts))
		inputContentPaths := make([]string, 0, len(dbOperator.Inputs))
		inputMetadataPaths := make([]string, 0, len(dbOperator.Inputs))
		for _, artifactID := range dbOperator.Inputs {
			inputArtifacts = append(inputArtifacts, artifacts[artifactID])
			inputContentPaths = append(inputContentPaths, artifactIDToContentPath[artifactID])
			inputMetadataPaths = append(inputMetadataPaths, artifactIDToContentPath[artifactID])

			artifactToOps[artifactID] = append(artifactToOps[artifactID], opID)
		}
		outputArtifacts := make([]artifact.Artifact, 0, len(artifacts))
		outputContentPaths := make([]string, 0, len(dbOperator.Outputs))
		outputMetadataPaths := make([]string, 0, len(dbOperator.Outputs))
		for _, artifactID := range dbOperator.Outputs {
			outputArtifacts = append(outputArtifacts, artifacts[artifactID])
			outputContentPaths = append(outputContentPaths, artifactIDToContentPath[artifactID])
			outputMetadataPaths = append(outputMetadataPaths, artifactIDToContentPath[artifactID])

			opToArtifacts[opID] = append(opToArtifacts[opID], artifactID)
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
			workflowDagResultID,
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

		dbWorkflowDag: dbWorkflowDag,
		operators:     operators,
		artifacts:     artifacts,
		opToArtifacts: opToArtifacts,
		artifactToOps: artifactToOps,

		workflowDagResultWriter: dagResultWriter,
		workflowReader:          workflowReader,
		notificationWriter:      notificationWriter,
		userReader:              userReader,
		db:                      db,

		// Can be nil, which means the dag cannot be persisted.
		workflowDagResultID: workflowDagResultID,
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

func (w *workflowDagImpl) ArtifactsFromOperator(op operator.Operator) ([]artifact.Artifact, error) {
	artifactIDs, ok := w.artifactToOps[op.ID()]
	if !ok {
		return nil, errors.Newf("Unable to find operator %s (%s) on dag.", op.ID(), op.Name())
	}

	artifacts := make([]artifact.Artifact, 0, len(artifactIDs))
	for _, artifactID := range artifactIDs {
		artifacts = append(artifacts, w.artifacts[artifactID])
	}
	return artifacts, nil
}

// Updates the dag result metadata after the dag has been executed.
// No-ops unless the dag result has already been initialized. This is meant to be called from a defer().
func (w *workflowDagImpl) PersistResult(ctx context.Context, status shared.ExecutionStatus) error {
	if w.workflowDagResultID == uuid.Nil {
		return errors.New("Cannot persist this workflow dag result. Initialized with `CanPersist` == false.")
	}

	// We `defer` this call to ensure that the WorkflowDagResult metadata is always updated.
	utils.UpdateWorkflowDagResultMetadata(
		ctx,
		w.workflowDagResultID,
		status,
		w.workflowDagResultWriter,
		w.workflowReader,
		w.notificationWriter,
		w.userReader,
		w.db,
	)

	return nil
}

func (w *workflowDagImpl) Finish(ctx context.Context) {
	for _, op := range w.operators {
		op.Finish(ctx)
	}
	for _, artf := range w.artifacts {
		artf.Finish(ctx)
	}
}
