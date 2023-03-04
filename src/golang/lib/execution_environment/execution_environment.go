package execution_environment

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"sort"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/lib_utils"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type ExecutionEnvironment struct {
	// TODO: Double check if the json tags can be removed.
	ID            uuid.UUID `json:"id"`
	PythonVersion string    `json:"python_version"`
	Dependencies  []string  `json:"dependencies"`
}

func (e *ExecutionEnvironment) CreateDBRecord(
	ctx context.Context,
	execEnvRepo repos.ExecutionEnvironment,
	db database.Database,
) error {
	hash, err := e.Hash()
	if err != nil {
		return err
	}

	dbEnv, err := execEnvRepo.Create(
		ctx,
		&shared.ExecutionEnvironmentSpec{
			PythonVersion: e.PythonVersion,
			Dependencies:  e.Dependencies,
		},
		hash,
		db,
	)
	if err != nil {
		return err
	}

	e.ID = dbEnv.ID
	return nil
}

func (e *ExecutionEnvironment) DeleteDBRecord(
	ctx context.Context,
	execEnvRepo repos.ExecutionEnvironment,
	db database.Database,
) error {
	return execEnvRepo.Delete(ctx, e.ID, db)
}

// Hash generates a hash based on the environment's
// dependency set and python version.
func (e *ExecutionEnvironment) Hash() (uuid.UUID, error) {
	sliceToHash := make([]string, 0, len(e.Dependencies)+1)
	sliceToHash = append(sliceToHash, e.Dependencies...)
	sliceToHash = append(sliceToHash, e.PythonVersion)
	sort.Strings(sliceToHash)

	buf := &bytes.Buffer{}
	err := gob.NewEncoder(buf).Encode(sliceToHash)
	if err != nil {
		return uuid.Nil, err
	}

	return uuid.NewSHA1(uuid.NameSpaceOID, buf.Bytes()), nil
}

func (e *ExecutionEnvironment) Name() string {
	return fmt.Sprintf("%s%s", aqueductEnvNamePrefix, e.ID.String())
}

// GetExecEnvFromDB returns an exec env object from DB by its hash.
// It returns database.ErrNoRows if there is no match.
func GetExecEnvFromDB(
	ctx context.Context,
	hash uuid.UUID,
	execEnvRepo repos.ExecutionEnvironment,
	db database.Database,
) (*ExecutionEnvironment, error) {
	dbExecEnv, err := execEnvRepo.GetByHash(ctx, hash, db)
	if err != nil {
		return nil, err
	}

	return newFromDBExecutionEnvironment(dbExecEnv), nil
}

func newFromDBExecutionEnvironment(
	dbExecEnv *models.ExecutionEnvironment,
) *ExecutionEnvironment {
	return &ExecutionEnvironment{
		ID:            dbExecEnv.ID,
		PythonVersion: dbExecEnv.Spec.PythonVersion,
		Dependencies:  dbExecEnv.Spec.Dependencies,
	}
}

func GetExecutionEnvironmentsByOperatorIDs(
	ctx context.Context,
	opIDs []uuid.UUID,
	execEnvRepo repos.ExecutionEnvironment,
	db database.Database,
) (map[uuid.UUID]ExecutionEnvironment, error) {
	dbEnvMap, err := execEnvRepo.GetByOperatorBatch(
		ctx, opIDs, db,
	)
	if err != nil {
		return nil, err
	}

	results := make(map[uuid.UUID]ExecutionEnvironment, len(dbEnvMap))
	for id, dbEnv := range dbEnvMap {
		results[id] = *newFromDBExecutionEnvironment(&dbEnv)
	}

	return results, nil
}

// CreateMissingAndSyncExistingEnvs keeps the given environment map in-sync with DB.
//
// Given a env map of any UUID key (in practice the key is typically operator ID),
// it creates new db rows for missing envs.
//
// Returns a map with the original key, mapped to the synced
// env object from the DB rows.
func CreateMissingAndSyncExistingEnvs(
	ctx context.Context,
	execEnvRepo repos.ExecutionEnvironment,
	envs map[uuid.UUID]ExecutionEnvironment,
	db database.Database,
) (map[uuid.UUID]ExecutionEnvironment, error) {
	// visitedResults tracks already visited env.
	// This helps reduce the number of DB access.
	visitedResults := make(map[uuid.UUID]ExecutionEnvironment, len(envs))
	results := make(map[uuid.UUID]ExecutionEnvironment, len(envs))
	var err error = nil
	txn, err := db.BeginTx(ctx)
	if err != nil {
		return nil, err
	}

	// rollback both DB records and conda envs.
	defer func() {
		database.TxnRollbackIgnoreErr(ctx, txn)
	}()

	for key, env := range envs {
		hash, err := env.Hash()
		if err != nil {
			return nil, err
		}

		_, ok := visitedResults[hash]
		if ok {
			results[key] = visitedResults[hash]
			continue
		}

		existingEnv, err := GetExecEnvFromDB(
			ctx,
			hash,
			execEnvRepo,
			txn,
		)

		// Env is missing
		if err == database.ErrNoRows {
			err = env.CreateDBRecord(ctx, execEnvRepo, txn)
			if err != nil {
				return nil, err
			}

			results[key] = env
			visitedResults[hash] = env
			continue
		}

		// DB error
		if err != nil {
			return nil, err
		}

		// Env is not missing
		visitedResults[hash] = *existingEnv
		results[key] = *existingEnv
	}

	if err = txn.Commit(ctx); err != nil {
		return nil, err
	}

	return results, nil
}

// CleanupUnusedEnvironments is executed in a best-effort fashion, and we log all the errors within
// the function and return an error object signaling whether there is at least one error occurred.
func CleanupUnusedEnvironments(
	ctx context.Context,
	operatorRepo repos.Operator,
	db database.Database,
) error {
	envNames, err := operatorRepo.GetUnusedCondaEnvNames(ctx, db)
	if err != nil {
		log.Errorf("Error getting unused execution environments: %v", err)
		return err
	}

	hasError := false

	for _, name := range envNames {
		deleteArgs := []string{
			"env",
			"remove",
			"-n",
			name,
		}

		_, _, err := lib_utils.RunCmd(CondaCmdPrefix, deleteArgs...)
		if err != nil {
			hasError = true
			log.Errorf("Error garbage collecting conda environment %s: %v", name, err)
		}
	}

	if hasError {
		return errors.New("An internal error occurred within the cleanup function. Please see the server log for more information.")
	}

	return nil
}

// MergeOperatorEnv merges a set of operator envs to generate
// one single workflow Env. This is only used for spark engine
// for now.
func MergeOperatorEnv(
	ctx context.Context,
	operatorEnvs map[uuid.UUID]ExecutionEnvironment,
	execEnvRepo repos.ExecutionEnvironment,
	DB database.Database,
) (*ExecutionEnvironment, error) {
	combinedDependenciesMap := make(map[string]bool)
	pythonVersion := ""
	for _, env := range operatorEnvs {
		for _, dep := range env.Dependencies {
			combinedDependenciesMap[dep] = true
		}

		if pythonVersion != "" && pythonVersion != env.PythonVersion {
			return nil, errors.New("Multiple python versions provided for different operators.")
		} else {
			pythonVersion = env.PythonVersion
		}
	}

	if pythonVersion == "" {
		pythonVersion = "3.9"
	}

	workflowDependencies := make([]string, 0, len(combinedDependenciesMap))
	for dep := range combinedDependenciesMap {
		workflowDependencies = append(workflowDependencies, dep)
	}
	workflowDependencies = append(workflowDependencies, "snowflake-sqlalchemy")

	workflowEnv := ExecutionEnvironment{ID: uuid.New()}
	workflowEnv.Dependencies = workflowDependencies
	workflowEnv.PythonVersion = pythonVersion
	envs, err := CreateMissingAndSyncExistingEnvs(
		ctx,
		execEnvRepo,
		map[uuid.UUID]ExecutionEnvironment{
			workflowEnv.ID: workflowEnv,
		},
		DB,
	)
	if err != nil {
		return nil, err
	}

	// CreateMissingAndSyncExistingEnvs preserves the original Key.
	env := envs[workflowEnv.ID]
	return &env, nil
}
