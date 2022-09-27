package dag

import (
	"context"
	db_artifact "github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact_result"
	"github.com/aqueducthq/aqueduct/lib/collections/notification"
	db_operator "github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/collections/operator_result"
	"github.com/aqueducthq/aqueduct/lib/collections/user"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_result"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/artifact"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator"
	"github.com/aqueducthq/aqueduct/lib/workflow/preview_cache"
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

	// InitOpAndArtifactResults initializes the operators and artifact results for this dag.
	InitOpAndArtifactResults(ctx context.Context) error
}

type workflowDagImpl struct {
	dbWorkflowDag *workflow_dag.DBWorkflowDag
	// resultID corresponds to the WorkflowDagResult created for the current run of this dag
	resultID uuid.UUID

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
}

// Assumption: all dag's start with operators.
// computeArtifactSignatures traverses over the entire dag structure from beginning to end,
// computing the signatures for each artifact. These signatures are returned in a map keyed
// by the artifact's original ID.
func computeArtifactSignatures(
	dbOperators map[uuid.UUID]db_operator.DBOperator,
	opIDsByInputArtifact map[uuid.UUID][]uuid.UUID,
	numArtifacts int,
) (map[uuid.UUID]uuid.UUID, error) {
	artifactIDToSignature := make(map[uuid.UUID]uuid.UUID, numArtifacts)

	// Queue that stores the frontier of operators as we perform a BFS over the dag.
	q := make([]uuid.UUID, 0, 1)
	for _, dbOperator := range dbOperators {
		if len(dbOperator.Inputs) == 0 {
			q = append(q, dbOperator.Id)
		}
	}

	processedArtifactIds := make(map[uuid.UUID]bool, numArtifacts)
	for len(q) > 0 {
		// Pop the first operator off of the queue.
		currOp := dbOperators[q[0]]
		q = q[1:]

		// Skip operators with no output artifacts.
		if len(currOp.Outputs) == 0 {
			continue
		}

		bytesToHash := []byte{}
		for _, inputArtifactID := range currOp.Inputs {
			inputArtifactSignature, ok := artifactIDToSignature[inputArtifactID]
			if !ok {
				return nil, errors.Newf("Unable to find signature for input artifact %s", inputArtifactID)
			}
			bytesToHash = append(bytesToHash, []byte(inputArtifactSignature.String())...)
		}

		// If the operator produces a parameter artifact, we also need to hash against the parameterized value.
		if currOp.Spec.Type() == db_operator.ParamType {
			bytesToHash = append(bytesToHash, []byte(currOp.Spec.Param().Val)...)
		}

		// Compute that signature for each output artifact.
		// The assumption is that there is only one output.
		for _, outputArtifactID := range currOp.Outputs {
			bytesToHash = append(bytesToHash, []byte(outputArtifactID.String())...)

			// Compute that final hash and add it to the map, then continue traversing.
			artifactIDToSignature[outputArtifactID] = uuid.NewSHA1(uuid.NameSpaceOID, bytesToHash)
			processedArtifactIds[outputArtifactID] = true

			// Find the next downstream operators that consume this output artifact.
			// In order to process it next, we must have already visited all the operator's inputs.
			for _, nextOpID := range opIDsByInputArtifact[outputArtifactID] {
				nextOp := dbOperators[nextOpID]

				depsComputed := true
				for _, inputArtifactID := range nextOp.Inputs {
					if _, ok := processedArtifactIds[inputArtifactID]; !ok {
						depsComputed = false
						break
					}
				}

				if depsComputed {
					q = append(q, nextOpID)
				}
			}
		}
	}
	return artifactIDToSignature, nil
}

func NewWorkflowDag(
	ctx context.Context,
	workflowDagResultID uuid.UUID,
	dbWorkflowDag *workflow_dag.DBWorkflowDag,
	dagResultWriter workflow_dag_result.Writer,
	opResultWriter operator_result.Writer,
	artifactWriter db_artifact.Writer,
	artifactResultWriter artifact_result.Writer,
	workflowReader workflow.Reader,
	notificationWriter notification.Writer,
	userReader user.Reader,
	jobManager job.JobManager,
	vaultObject vault.Vault,
	artifactCacheManager preview_cache.CacheManager,
	opExecMode operator.ExecutionMode,
	db database.Database,
) (WorkflowDag, error) {
	dbArtifacts := dbWorkflowDag.Artifacts
	dbOperators := dbWorkflowDag.Operators

	artifactIDToInputOpID := make(map[uuid.UUID]uuid.UUID, len(dbArtifacts))
	opIDToMetadataPath := make(map[uuid.UUID]string, len(dbOperators))
	opIDsByInputArtifact := make(map[uuid.UUID][]uuid.UUID, len(dbArtifacts))
	for _, dbOperator := range dbOperators {
		for _, outputArtifactID := range dbOperator.Outputs {
			artifactIDToInputOpID[outputArtifactID] = dbOperator.Id
		}
		opIDToMetadataPath[dbOperator.Id] = utils.InitializePath(opExecMode == operator.Preview)

		for _, inputArtifactID := range dbOperator.Inputs {
			opIDs, ok := opIDsByInputArtifact[inputArtifactID]
			if !ok {
				opIDs = make([]uuid.UUID, 0, 1)
			}
			opIDsByInputArtifact[inputArtifactID] = append(opIDs, dbOperator.Id)
		}
	}

	// Allocate all execution paths for the workflowlib/workflow/operator/base.go.
	artifactIDToExecPaths := make(map[uuid.UUID]*utils.ExecPaths, len(dbArtifacts))
	for _, dbArtifact := range dbArtifacts {
		inputOpID := artifactIDToInputOpID[dbArtifact.Id]
		opMetadataPath, ok := opIDToMetadataPath[inputOpID]
		if !ok {
			return nil, errors.Newf("DAGs cannot currently start with an artifact.")
		}

		artifactIDToExecPaths[dbArtifact.Id] = utils.InitializeExecOutputPaths(
			opExecMode == operator.Preview,
			opMetadataPath,
		)
	}

	// Compute signatures for each artifact.
	artifactIDToSignatures, err := computeArtifactSignatures(dbOperators, opIDsByInputArtifact, len(dbArtifacts))
	if err != nil {
		return nil, errors.Wrap(err, "Internal error: unable to set up workflow execution.")
	}

	// With all the initial database writes completed (if at all), we can now initialize
	// the operator and artifact classes. As well as the connections between them.
	operators := make(map[uuid.UUID]operator.Operator, len(dbWorkflowDag.Operators))
	artifacts := make(map[uuid.UUID]artifact.Artifact, len(dbWorkflowDag.Artifacts))
	for artifactID, dbArtifact := range dbWorkflowDag.Artifacts {
		newArtifact, err := artifact.NewArtifact(
			artifactIDToSignatures[dbArtifact.Id],
			dbArtifact,
			artifactIDToExecPaths[artifactID],
			artifactWriter,
			artifactResultWriter,
			&dbWorkflowDag.StorageConfig,
			artifactCacheManager,
			db,
		)
		if err != nil {
			return nil, err
		}
		artifacts[artifactID] = newArtifact
	}

	// These artifact <-> operator maps help us remember all dag connections.
	artifactIDToOpIDs := make(map[uuid.UUID][]uuid.UUID, len(dbArtifacts))
	for _, dbArtifact := range dbArtifacts {
		artifactIDToOpIDs[dbArtifact.Id] = make([]uuid.UUID, 0, 1)
	}
	opToInputArtifactIDs := make(map[uuid.UUID][]uuid.UUID, len(dbOperators))
	opToOutputArtifactIDs := make(map[uuid.UUID][]uuid.UUID, len(dbOperators))
	for opID, dbOperator := range dbOperators {
		opToOutputArtifactIDs[opID] = make([]uuid.UUID, 0, 1)
		opToInputArtifactIDs[opID] = make([]uuid.UUID, 0, 1)

		inputArtifacts := make([]artifact.Artifact, 0, len(dbOperator.Inputs))
		inputExecPaths := make([]*utils.ExecPaths, 0, len(dbOperator.Inputs))
		for _, artifactID := range dbOperator.Inputs {
			inputArtifacts = append(inputArtifacts, artifacts[artifactID])
			inputExecPaths = append(inputExecPaths, artifactIDToExecPaths[artifactID])

			artifactIDToOpIDs[artifactID] = append(artifactIDToOpIDs[artifactID], opID)
			opToInputArtifactIDs[opID] = append(opToInputArtifactIDs[opID], artifactID)
		}

		outputArtifacts := make([]artifact.Artifact, 0, len(dbOperator.Outputs))
		outputExecPaths := make([]*utils.ExecPaths, 0, len(dbOperator.Outputs))
		for _, artifactID := range dbOperator.Outputs {
			outputArtifacts = append(outputArtifacts, artifacts[artifactID])
			outputExecPaths = append(outputExecPaths, artifactIDToExecPaths[artifactID])

			opToOutputArtifactIDs[opID] = append(opToOutputArtifactIDs[opID], artifactID)
		}

		newOp, err := operator.NewOperator(
			ctx,
			dbOperator,
			inputArtifacts,
			outputArtifacts,
			inputExecPaths,
			outputExecPaths,
			opResultWriter,
			jobManager,
			vaultObject,
			&dbWorkflowDag.StorageConfig,
			artifactCacheManager,
			opExecMode,
			db,
		)
		if err != nil {
			return nil, err
		}
		operators[opID] = newOp
	}

	return &workflowDagImpl{
		dbWorkflowDag:       dbWorkflowDag,
		resultID:            workflowDagResultID,
		operators:           operators,
		artifacts:           artifacts,
		opToOutputArtifacts: opToOutputArtifactIDs,
		opToInputArtifacts:  opToInputArtifactIDs,
		artifactToOps:       artifactIDToOpIDs,

		resultWriter:       dagResultWriter,
		workflowReader:     workflowReader,
		notificationWriter: notificationWriter,
		userReader:         userReader,
		db:                 db,
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

func (w *workflowDagImpl) InitOpAndArtifactResults(ctx context.Context) error {
	// Initialize the operators and artifact results.
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
