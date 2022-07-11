package dag

import (
"context"
"github.com/aqueducthq/aqueduct/cmd/server/request"
"github.com/aqueducthq/aqueduct/lib/collections/artifact_result"
"github.com/aqueducthq/aqueduct/lib/collections/operator_result"
"github.com/aqueducthq/aqueduct/lib/collections/shared"
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
log "github.com/sirupsen/logrus"
)

type WorkflowDag struct {
	DBWorkflowDag *workflow_dag.DBWorkflowDag

	Operators map[uuid.UUID]operator.Operator
	Artifacts map[uuid.UUID]artifact.Artifact

	isPreview               bool
	workflowDagResultWriter workflow_dag_result.Writer

	// A convenience data structure mapping artifacts to the operators that consumes it as input.
	opsByInputArtifact map[uuid.UUID][]operator.Operator
}

func initializeDagResultInDatabase(
	ctx context.Context,
	dbWorkflowDag *workflow_dag.DBWorkflowDag,
	workflowDagResultWriter workflow_dag_result.Writer,
	db database.Database,
) (uuid.UUID, error) {
	// Create a database record of workflow dag result and set its status to `pending`.
	// TODO(ENG-599): wrap these writes into a transaction.
	workflowDagResult, err := workflowDagResultWriter.CreateWorkflowDagResult(ctx, dbWorkflowDag.Id, db)
	if err != nil {
		return uuid.Nil, errors.Wrap(err, "Unable to create workflow dag result record.")
	}
	return workflowDagResult.Id, nil
}

func NewWorkflowDag(
	ctx context.Context,
	dagSummary *request.DagSummary,
	workflowDagResultWriter workflow_dag_result.Writer,
	opResultWriter operator_result.Writer,
	artifactResultWriter artifact_result.Writer,
	jobManager job.JobManager,
	vaultObject vault.Vault,
	storageConfig *shared.StorageConfig,
	db database.Database,
) (*WorkflowDag, error) {
	dbWorkflowDag := dagSummary.Dag
	isPreview := workflowDagResultWriter != nil && opResultWriter != nil && artifactResultWriter != nil

	// First, allocate a content and metadata path for each artifact.
	artifactIDToContentPath := make(map[uuid.UUID]string, len(dbWorkflowDag.Artifacts))
	artifactIDToMetadataPath := make(map[uuid.UUID]string, len(dbWorkflowDag.Artifacts))
	for _, dbArtifact := range dbWorkflowDag.Artifacts {
		artifactIDToContentPath[dbArtifact.Id] = uuid.New().String()
		artifactIDToMetadataPath[dbArtifact.Id] = uuid.New().String()
	}

	var workflowDagResultID uuid.UUID
	var err error
	if !isPreview {
		workflowDagResultID, err = initializeDagResultInDatabase(
			ctx,
			dbWorkflowDag,
			workflowDagResultWriter,
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

	for opID, dbOperator := range dbWorkflowDag.Operators {
		inputArtifacts := make([]artifact.Artifact, 0, len(artifacts))
		inputContentPaths := make([]string, 0, len(dbOperator.Inputs))
		inputMetadataPaths := make([]string, 0, len(dbOperator.Inputs))
		for _, artifactID := range dbOperator.Inputs {
			inputArtifacts = append(inputArtifacts, artifacts[artifactID])
			inputContentPaths = append(inputContentPaths, artifactIDToContentPath[artifactID])
			inputMetadataPaths = append(inputMetadataPaths, artifactIDToContentPath[artifactID])
		}
		outputArtifacts := make([]artifact.Artifact, 0, len(artifacts))
		outputContentPaths := make([]string, 0, len(dbOperator.Outputs))
		outputMetadataPaths := make([]string, 0, len(dbOperator.Outputs))
		for _, artifactID := range dbOperator.Outputs {
			outputArtifacts = append(outputArtifacts, artifacts[artifactID])
			outputContentPaths = append(outputContentPaths, artifactIDToContentPath[artifactID])
			outputMetadataPaths = append(outputMetadataPaths, artifactIDToContentPath[artifactID])
		}

		operators[opID], err = operator.NewOperator(
			ctx,
			dbOperator,
			inputArtifacts,
			inputContentPaths,
			inputMetadataPaths,
			outputArtifacts,gg
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

	opsByInputArtifact := make(map[uuid.UUID][]operator.Operator, len(artifacts))
	for _, op := range operators {
		for _, inputArtifact := range op.Inputs() {
			ops, ok := opsByInputArtifact[inputArtifact.ID()]
			if !ok {
				ops = make([]operator.Operator, 0, 1)
			}
			opsByInputArtifact[inputArtifact.ID()] = append(ops, op)
		}
	}

	return &WorkflowDag{
		DBWorkflowDag: dbWorkflowDag,
		Operators:     operators,
		Artifacts:     artifacts,

		// The following fields are internal to the class methods only.
		isPreview:               isPreview,
		workflowDagResultWriter: workflowDagResultWriter,
		opsByInputArtifact:      opsByInputArtifact,
	}, nil
}

func (w *WorkflowDag) ImmediateDownstreamOperators(op operator.Operator) []operator.Operator {
	ops := make([]operator.Operator, 0, len(w.Operators))
	for _, outputArtifact := range op.Outputs() {
		nextOps := w.opsByInputArtifact[outputArtifact.ID()]
		ops = append(ops, nextOps...)
	}
	return ops
}

// - Update the workflow dag result metadata.
// This is meant to be called from a defer().
func (w *WorkflowDag) Flush() {
	if w.isPreview {
		log.Errorf("Flush() was called on workflow that was not supposed to write anything.")
		return
	}

	// We `defer` this call to ensure that the WorkflowDagResult metadata is always updated.
	utils.UpdateWorkflowDagResultMetadata(
		ctx,
		workflowDagResultId,
		status,
		workflowDagResultWriter,
		workflowReader,
		notificationWriter,
		userReader,
		db,
	)
}