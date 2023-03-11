package operator

import (
	"context"
	"time"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/errors"
	exec_env "github.com/aqueducthq/aqueduct/lib/execution_environment"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/lib_utils"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/models/shared/operator"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/artifact"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	"github.com/aqueducthq/aqueduct/lib/workflow/preview_cache"
	"github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/google/uuid"
)

// Operator is an interface for managing and inspecting the lifecycle of an operator
// used by a workflow run.
type Operator interface {
	// Property getters. Retrieve property of the operator without making any changes.
	Type() operator.Type
	Name() string
	ID() uuid.UUID
	JobSpec() job.Spec
	// ExecState returns the operators ExecState since the last `Poll()` or `Launch()`
	ExecState() *shared.ExecutionState

	// Execution methods. These methods trigger or interact with jobs
	// and update operator's state.
	// However, they do not persist anything to DB.

	// Launch kicks off the execution of this operator, using operator's job spec.
	// It sets the operator's execState to 'Running' without updating DB.
	// Use `Poll()` afterwards to determine when this operator has completed.
	Launch(ctx context.Context) error

	// Poll performs a non-blocking fetch and update for the execution state of this operator.
	// Returns the execState updated. This does not persist the exec state to DB.
	Poll(ctx context.Context) (*shared.ExecutionState, error)

	// Cancel updates the status of this operator execution if the result of the
	// execution will not be generated. This does not persist the exec state to DB.
	Cancel()

	// Finish is an end-of-lifecycle hook meant to do any final cleanup work.
	// Also calls Finish() on all the operator's output artifacts.
	Finish(ctx context.Context)

	// DB methods, these methods update DB based on current operator's state.

	// InitializeResult initializes the operator in the database.
	InitializeResult(ctx context.Context, dagResultID uuid.UUID) error

	// PersistResult writes the results of this operator execution to the database.
	// The result persisted is based on the last `Poll()`.
	//
	// Errors if the artifact hasn ot yet been computed, or InitializeResult() hasn't been called yet.
	// *This method also persists any artifact results produced by this operator.*
	PersistResult(ctx context.Context) error

	// FetchExecState retrieves the execution state from storage.
	FetchExecState(ctx context.Context) *shared.ExecutionState

	// UpdateExecState and merge timestamps with current state based on the status of the new state.
	// Other fields of bo.execState will be replaced.
	UpdateExecState(execState *shared.ExecutionState)

	// Dynamic returns a bool indicating whether an operator is using a dynamic engine.
	Dynamic() bool

	// GetDynamicProperties retrieves the dynamic properties of an operator, which includes its
	// engine integration ID and its `prepared` flag.
	GetDynamicProperties() *dynamicProperties
	// FetchExecutionEnvironment retrieves the environment of the operator.
	FetchExecutionEnvironment(ctx context.Context) *exec_env.ExecutionEnvironment
}

// This should only be used within the boundaries of the execution engine.
// Specifies what we will do with the operator's results.
// Preview: *does not* persist workflow results or write to third-party integrations.
// Publish *does* both.
type ExecutionMode int

const (
	Preview ExecutionMode = iota
	Publish
)

func NewOperator(
	ctx context.Context,
	dbOperator models.Operator,
	inputs []artifact.Artifact,
	outputs []artifact.Artifact,
	inputExecPaths []*utils.ExecPaths,
	outputExecPaths []*utils.ExecPaths,
	opResultRepo repos.OperatorResult, // A nil value means the operator is run in preview mode.
	opEngineConfig shared.EngineConfig,
	vaultObject vault.Vault,
	storageConfig *shared.StorageConfig,
	previewCacheManager preview_cache.CacheManager,
	execMode ExecutionMode,
	execEnv *exec_env.ExecutionEnvironment,
	aqPath string,
	db database.Database,
	dagJobManager job.JobManager, // Override that is only used when operator jobManagers need shared context.
) (Operator, error) {
	if len(inputs) != len(inputExecPaths) {
		return nil, errors.New("Internal error: mismatched number of input arguments.")
	}

	if len(outputs) != len(outputExecPaths) {
		return nil, errors.New("Internal error: mismatched number of input arguments.")
	}

	// If this operator has no outputs, we will need to allocate a new metadata path.
	// This is because the operator's metadata path is defined on the operator's outputs.
	metadataPath := uuid.New().String()
	if len(outputExecPaths) > 0 {
		metadataPath = outputExecPaths[0].OpMetadataPath
	}

	var jobManager job.JobManager
	var err error

	if dagJobManager == nil {
		jobManager, err = job.GenerateNewJobManager(
			ctx, opEngineConfig, storageConfig, aqPath, vaultObject,
		)
		if err != nil {
			return nil, err
		}
	} else {
		jobManager = dagJobManager
	}

	var dProperties *dynamicProperties

	if opEngineConfig.Type == shared.K8sEngineType {
		k8sIntegrationId := opEngineConfig.K8sConfig.IntegrationID
		config, err := auth.ReadConfigFromSecret(ctx, k8sIntegrationId, vaultObject)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to read k8s config from vault.")
		}
		k8sConfig, err := lib_utils.ParseK8sConfig(config)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to parse k8s config.")
		}

		if k8sConfig.Dynamic {
			dProperties = &dynamicProperties{
				engineIntegrationId: k8sIntegrationId,
				prepared:            false,
			}
		}
	}

	now := time.Now()

	baseOp := baseOperator{
		dbOperator: &dbOperator,
		resultRepo: opResultRepo,
		resultID:   uuid.Nil,

		metadataPath: metadataPath,
		jobName:      "", /* Must be set by the specific type constructors below. */

		inputs:          inputs,
		outputs:         outputs,
		inputExecPaths:  inputExecPaths,
		outputExecPaths: outputExecPaths,

		previewCacheManager: previewCacheManager,
		jobManager:          jobManager,
		vaultObject:         vaultObject,
		storageConfig:       storageConfig,
		db:                  db,

		execMode: execMode,
		execState: shared.ExecutionState{
			Status: shared.PendingExecutionStatus,
			Timestamps: &shared.ExecutionTimestamps{
				PendingAt: &now,
			},
		},

		// These fields may be set dynamically during orchestration.
		resultsPersisted:  false,
		execEnv:           execEnv,
		dynamicProperties: dProperties,
	}

	if dbOperator.Spec.IsFunction() {
		return newFunctionOperator(baseFunctionOperator{baseOp})
	} else if dbOperator.Spec.IsMetric() {
		return newMetricOperator(baseFunctionOperator{baseOp})
	} else if dbOperator.Spec.IsCheck() {
		return newCheckOperator(baseFunctionOperator{baseOp})
	} else if dbOperator.Spec.IsExtract() {
		return newExtractOperator(ctx, baseOp)
	} else if dbOperator.Spec.IsLoad() {
		return newLoadOperator(ctx, baseOp)
	} else if dbOperator.Spec.IsParam() {
		return newParamOperator(baseOp)
	} else if dbOperator.Spec.IsSystemMetric() {
		return newSystemMetricOperator(baseOp)
	}

	return nil, errors.Newf("Unsupported operator type %s", dbOperator.Spec.Type())
}
