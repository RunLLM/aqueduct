package execution_environment

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"sort"

	"github.com/aqueducthq/aqueduct/lib"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/lib_utils"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

var pythonVersions = [...]string{
	"3.7",
	"3.8",
	"3.9",
	"3.10",
}

type ExecutionEnvironment struct {
	// TODO: Double check if the json tags can be removed.
	ID            uuid.UUID `json:"id"`
	PythonVersion string    `json:"python_version"`
	Dependencies  []string  `json:"dependencies"`
	CondaPath     string    `json:"-"`
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
	sort.Strings(e.Dependencies)

	buf := &bytes.Buffer{}
	err := gob.NewEncoder(buf).Encode(sliceToHash)
	if err != nil {
		return uuid.Nil, err
	}

	return uuid.NewSHA1(uuid.NameSpaceOID, buf.Bytes()), nil
}

func (e *ExecutionEnvironment) Name() string {
	return fmt.Sprintf("aqueduct_%s", e.ID.String())
}

func (e *ExecutionEnvironment) CreateCondaEnv() error {
	// First, we create a conda env with the env object's Python version.
	createArgs := []string{
		"create",
		"-n",
		e.Name(),
		fmt.Sprintf("python==%s", e.PythonVersion),
		"-y",
	}

	_, _, err := lib_utils.RunCmd(CondaCmdPrefix, createArgs...)
	if err != nil {
		return err
	}

	forkEnvPath := fmt.Sprintf(
		"%s/envs/aqueduct_python%s/lib/python%s/site-packages",
		e.CondaPath,
		e.PythonVersion,
		e.PythonVersion,
	)
	forkArgs := []string{
		"develop",
		"-n",
		e.Name(),
		forkEnvPath,
	}

	_, _, err = lib_utils.RunCmd(CondaCmdPrefix, forkArgs...)
	if err != nil {
		return err
	}

	// Then, we use pip3 to install dependencies inside this new Conda env.
	installArgs := append([]string{
		"run",
		"-n",
		e.Name(),
		"pip3",
		"install",
	}, e.Dependencies...)

	_, _, err = lib_utils.RunCmd(CondaCmdPrefix, installArgs...)
	if err != nil {
		return err
	}

	return nil
}

func deleteCondaEnv(name string) error {
	args := []string{
		"env",
		"remove",
		"-n",
		name,
	}

	_, _, err := lib_utils.RunCmd(CondaCmdPrefix, args...)
	return err
}

// DeleteEnv deletes the Conda environment if it exists.
func (e *ExecutionEnvironment) DeleteCondaEnv() error {
	return deleteCondaEnv(e.Name())
}

// GetExecEnvFromDB returns an exec env object from DB by its hash.
// It returns database.ErrNoRows if there is no match.
func GetExecEnvFromDB(
	ctx context.Context,
	hash uuid.UUID,
	execEnvRepo repos.ExecutionEnvironment,
	db database.Database,
) (*ExecutionEnvironment, error) {
	dbExecEnv, err := execEnvRepo.GetActiveByHash(ctx, hash, db)
	if err != nil {
		return nil, err
	}

	return newFromDBExecutionEnvironment(dbExecEnv), nil
}

func baseEnvNameByVersion(pythonVersion string) string {
	return fmt.Sprintf("aqueduct_python%s", pythonVersion)
}

// createBaseEnvs creates base python environments.
func createBaseEnvs() error {
	for _, pythonVersion := range pythonVersions {
		envName := baseEnvNameByVersion(pythonVersion)
		args := []string{
			"create",
			"-n",
			envName,
			fmt.Sprintf("python==%s", pythonVersion),
			"-y",
		}
		_, _, err := lib_utils.RunCmd(CondaCmdPrefix, args...)
		if err != nil {
			return err
		}

		args = []string{
			"run",
			"-n",
			envName,
			"pip3",
			"install",
			fmt.Sprintf("aqueduct-ml==%s", lib.ServerVersionNumber),
		}
		_, _, err = lib_utils.RunCmd(CondaCmdPrefix, args...)
		if err != nil {
			return err
		}
	}

	return nil
}

func DeleteBaseEnvs() error {
	for _, pythonVersion := range pythonVersions {
		err := deleteCondaEnv(baseEnvNameByVersion(pythonVersion))
		if err != nil {
			return err
		}
	}

	return nil
}

// Best-effort to delete all envs and log any error
func deleteEnvs(envs []ExecutionEnvironment) {
	for _, env := range envs {
		err := env.DeleteEnv()
		if err != nil {
			log.Errorf("Failed to delete env %s: %v", env.ID.String(), err)
		}
	}
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

func GetActiveExecutionEnvironmentsByOperatorIDs(
	ctx context.Context,
	opIDs []uuid.UUID,
	execEnvRepo repos.ExecutionEnvironment,
	db database.Database,
) (map[uuid.UUID]ExecutionEnvironment, error) {
	dbEnvMap, err := execEnvRepo.GetActiveByOperatorBatch(
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
	addedEnvs := make([]ExecutionEnvironment, 0, len(envs))
	results := make(map[uuid.UUID]ExecutionEnvironment, len(envs))
	var err error = nil
	txn, err := db.BeginTx(ctx)
	if err != nil {
		return nil, err
	}

	// rollback both DB records and conda envs.
	defer func() {
		database.TxnRollbackIgnoreErr(ctx, txn)
		if err != nil {
			deleteEnvs(addedEnvs)
		}
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
			db,
		)

		// Env is missing
		if err == database.ErrNoRows {
			err = env.CreateDBRecord(ctx, execEnvRepo, db)
			if err != nil {
				return nil, err
			}

			err = env.CreateEnv()
			if err != nil {
				return nil, err
			}

			results[key] = env
			visitedResults[hash] = env
			addedEnvs = append(addedEnvs, env)
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

func GetUnusedExecutionEnvironmentIDs(
	ctx context.Context,
	execEnvRepo repos.ExecutionEnvironment,
	db database.Database,
) ([]uuid.UUID, error) {
	dbEnvs, err := execEnvRepo.GetUnused(
		ctx, db,
	)
	if err != nil {
		return nil, err
	}

	results := make([]uuid.UUID, 0, len(dbEnvs))
	for _, dbEnv := range dbEnvs {
		results = append(results, dbEnv.ID)
	}

	return results, nil
}

// CleanupUnusedEnvironments is executed in a best-effort fashion, and we log all the errors within
// the function and return an error object signaling whether there is at least one error occurred.
func CleanupUnusedEnvironments(
	ctx context.Context,
	execEnvRepo repos.ExecutionEnvironment,
	db database.Database,
) error {
	envIDs, err := GetUnusedExecutionEnvironmentIDs(ctx, execEnvRepo, db)
	if err != nil {
		log.Errorf("Error getting unused execution environments: %v", err)
		return err
	}

	hasError := false

	for _, envID := range envIDs {
		envName := fmt.Sprintf("%s_%s", "aqueduct", envID.String())
		deleteArgs := []string{
			"env",
			"remove",
			"-n",
			envName,
		}

		_, _, err := lib_utils.RunCmd(CondaCmdPrefix, deleteArgs...)
		if err != nil {
			hasError = true
			log.Errorf("Error garbage collecting conda environment %s: %v", envID, err)
		} else {
			_, err = execEnvRepo.Update(
				ctx,
				envID,
				map[string]interface{}{
					"garbage_collected": true,
				},
				db,
			)
			log.Errorf("Error updating the garbage collection column of conda environment %s: %v", envID, err)
		}
	}

	if hasError {
		return errors.New("An internal error occurred within the cleanup function. Please see the server log for more information.")
	}

	return nil
}
