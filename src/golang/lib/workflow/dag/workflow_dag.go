package dag

import (
	"context"

	db_operator "github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/database"
	exec_env "github.com/aqueducthq/aqueduct/lib/execution_environment"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/repos"
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

	// FindMissingExecEnv returns `Environment` objects for all missing environments
	// of all operators on this DAG.
	FindMissingExecEnv(ctx context.Context) ([]exec_env.ExecutionEnvironment, error)

	// BindOperatorsToEnvs updates all operators such that each operator
	// points to the environment object matching its dependencies.
	//
	// This function assumes there's no missing environments, and should error
	// if there's any environment missing.
	//
	// This function DO NOT update operators in DB. One should call `op.Persist()`
	// to do so.
	BindOperatorsToEnvs(ctx context.Context) error
}

type workflowDagImpl struct {
	dbDAG *models.DAG
	// resultID corresponds to the WorkflowDagResult created for the current run of this dag
	resultID uuid.UUID

	operators           map[uuid.UUID]operator.Operator
	artifacts           map[uuid.UUID]artifact.Artifact
	opToOutputArtifacts map[uuid.UUID][]uuid.UUID
	opToInputArtifacts  map[uuid.UUID][]uuid.UUID
	artifactToOps       map[uuid.UUID][]uuid.UUID
}

// Assumption: all dag's start with operators.
// computeArtifactSignatures traverses over the entire dag structure from beginning to end,
// computing the signatures for each artifact. These signatures are returned in a map keyed
// by the artifact's original ID.
// `opIDsByInputArtifact` does not contain entries for terminal artifacts.
func computeArtifactSignatures(
	dbOperators map[uuid.UUID]models.Operator,
	opIDsByInputArtifact map[uuid.UUID][]uuid.UUID,
	numArtifacts int,
) (map[uuid.UUID]uuid.UUID, error) {
	artifactIDToSignature := make(map[uuid.UUID]uuid.UUID, numArtifacts)

	// Queue that stores the frontier of operators as we perform a BFS over the dag.
	q := make([]uuid.UUID, 0, 1)
	for _, dbOperator := range dbOperators {
		if len(dbOperator.Inputs) == 0 {
			q = append(q, dbOperator.ID)
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

		// Represents the bytes prefix that we want to hash for each output artifact.
		// Is computed to be the concatenation of the operator's input signatures, along with
		// the parameter value of the operator (if the operator is a parameter).
		// These bytes are meant to be concatenated with each output artifact's id and the result
		// hashed in order to obtain the signature for each output artifact.
		inputBytesToHash := []byte{}
		for _, inputArtifactID := range currOp.Inputs {
			inputArtifactSignature, ok := artifactIDToSignature[inputArtifactID]
			if !ok {
				return nil, errors.Newf("Unable to find signature for input artifact %s", inputArtifactID)
			}
			inputBytesToHash = append(inputBytesToHash, []byte(inputArtifactSignature.String())...)
		}

		// If the operator produces a parameter artifact, we also need to hash against the parameterized value.
		if currOp.Spec.Type() == db_operator.ParamType {
			inputBytesToHash = append(inputBytesToHash, []byte(currOp.Spec.Param().Val)...)
		}

		// Compute that signature for each output artifact.
		for _, outputArtifactID := range currOp.Outputs {
			// NOTE: is it important for correctness that we do not allocate additional capacity for `inputBytesToHash`.
			// We need append() to always create a new slice for each `outputBytesToHash`. This allows us to use
			// `inputBytesToHash` multiple times within this loop without worrying about it changing.
			// From a performance standpoint, it can be suboptimal, since `inputBytesToHash` will be copied
			// to a new location on each append() call.
			outputBytesToHash := append(inputBytesToHash, []byte(outputArtifactID.String())...)

			// Compute that final hash and add it to the map, then continue traversing.
			artifactIDToSignature[outputArtifactID] = uuid.NewSHA1(uuid.NameSpaceOID, outputBytesToHash)
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
	dagResultID uuid.UUID,
	dag *models.DAG,
	opResultRepo repos.OperatorResult,
	artifactRepo repos.Artifact,
	artifactResultRepo repos.ArtifactResult,
	vaultObject vault.Vault,
	artifactCacheManager preview_cache.CacheManager,
	execEnvs map[uuid.UUID]exec_env.ExecutionEnvironment,
	opExecMode operator.ExecutionMode,
	aqPath string,
	DB database.Database,
) (WorkflowDag, error) {
	dbArtifacts := dag.Artifacts
	dbOperators := dag.Operators

	artifactIDToInputOpID := make(map[uuid.UUID]uuid.UUID, len(dbArtifacts))
	opIDToMetadataPath := make(map[uuid.UUID]string, len(dbOperators))
	opIDsByInputArtifact := make(map[uuid.UUID][]uuid.UUID, len(dbArtifacts))
	for _, dbOperator := range dbOperators {
		for _, outputArtifactID := range dbOperator.Outputs {
			artifactIDToInputOpID[outputArtifactID] = dbOperator.ID
		}
		opIDToMetadataPath[dbOperator.ID] = utils.InitializePath(opExecMode == operator.Preview)

		for _, inputArtifactID := range dbOperator.Inputs {
			opIDs, ok := opIDsByInputArtifact[inputArtifactID]
			if !ok {
				opIDs = make([]uuid.UUID, 0, 1)
			}
			opIDsByInputArtifact[inputArtifactID] = append(opIDs, dbOperator.ID)
		}
	}

	// Allocate all execution paths for the workflowlib/workflow/operator/base.go.
	artifactIDToExecPaths := make(map[uuid.UUID]*utils.ExecPaths, len(dbArtifacts))
	for _, dbArtifact := range dbArtifacts {
		inputOpID := artifactIDToInputOpID[dbArtifact.ID]
		opMetadataPath, ok := opIDToMetadataPath[inputOpID]
		if !ok {
			return nil, errors.Newf("DAGs cannot currently start with an artifact.")
		}

		artifactIDToExecPaths[dbArtifact.ID] = utils.InitializeExecOutputPaths(
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
	operators := make(map[uuid.UUID]operator.Operator, len(dag.Operators))
	artifacts := make(map[uuid.UUID]artifact.Artifact, len(dag.Artifacts))
	for artifactID, dbArtifact := range dag.Artifacts {
		newArtifact, err := artifact.NewArtifact(
			artifactIDToSignatures[dbArtifact.ID],
			dbArtifact,
			artifactIDToExecPaths[artifactID],
			artifactRepo,
			artifactResultRepo,
			&dag.StorageConfig,
			artifactCacheManager,
			DB,
		)
		if err != nil {
			return nil, err
		}
		artifacts[artifactID] = newArtifact
	}

	// These artifact <-> operator maps help us remember all dag connections.
	artifactIDToOpIDs := make(map[uuid.UUID][]uuid.UUID, len(dbArtifacts))
	for _, dbArtifact := range dbArtifacts {
		artifactIDToOpIDs[dbArtifact.ID] = make([]uuid.UUID, 0, 1)
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

		execEnv, ok := execEnvs[opID]
		var execEnvPtr *exec_env.ExecutionEnvironment = nil
		if ok {
			execEnvPtr = &execEnv
		}

		// Operator's engine takes precedence over dag's engine.
		opEngineConfig := dag.EngineConfig
		if dbOperator.Spec.EngineConfig() != nil {
			opEngineConfig = *dbOperator.Spec.EngineConfig()
		}

		newOp, err := operator.NewOperator(
			ctx,
			dbOperator,
			inputArtifacts,
			outputArtifacts,
			inputExecPaths,
			outputExecPaths,
			opResultRepo,
			opEngineConfig,
			vaultObject,
			&dag.StorageConfig,
			artifactCacheManager,
			opExecMode,
			execEnvPtr,
			aqPath,
			DB,
		)
		if err != nil {
			return nil, err
		}
		operators[opID] = newOp
	}

	return &workflowDagImpl{
		dbDAG:               dag,
		resultID:            dagResultID,
		operators:           operators,
		artifacts:           artifacts,
		opToOutputArtifacts: opToOutputArtifactIDs,
		opToInputArtifacts:  opToInputArtifactIDs,
		artifactToOps:       artifactIDToOpIDs,
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
