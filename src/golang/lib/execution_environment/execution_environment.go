package execution_environment

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"sort"
	"time"

	"github.com/aqueducthq/aqueduct/lib"
	db_exec_env "github.com/aqueducthq/aqueduct/lib/collections/execution_environment"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/lib_utils"
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
	Id            uuid.UUID `json:"id"`
	PythonVersion string    `json:"python_version"`
	Dependencies  []string  `json:"dependencies"`
	CondaPath     string    `json:"-"`
}

func (e *ExecutionEnvironment) CreateDBRecord(
	ctx context.Context,
	execEnvWriter db_exec_env.Writer,
	db database.Database,
) error {
	hash, err := e.Hash()
	if err != nil {
		return err
	}

	dbEnv, err := execEnvWriter.CreateExecutionEnvironment(
		ctx,
		&db_exec_env.Spec{
			PythonVersion: e.PythonVersion,
			Dependencies:  e.Dependencies,
		},
		hash,
		db,
	)
	if err != nil {
		return err
	}

	e.Id = dbEnv.Id
	return nil
}

func (e *ExecutionEnvironment) DeleteDBRecord(
	ctx context.Context,
	execEnvWriter db_exec_env.Writer,
	db database.Database,
) error {
	return execEnvWriter.DeleteExecutionEnvironment(ctx, e.Id, db)
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
	return fmt.Sprintf("aqueduct_%s", e.Id.String())
}

func (e *ExecutionEnvironment) CreateEnv() error {
	// First, we create a Conda env with the env object's Python version.
	createArgs := []string{
		"create",
		"-n",
		e.Name(),
		fmt.Sprintf("python==%s", e.PythonVersion),
		"-y",
	}

	start := time.Now()
	_, _, err := lib_utils.RunCmd(CondaCmdPrefix, createArgs...)
	if err != nil {
		return err
	}
	duration := time.Since(start)
	log.Infof("Conda creation took %d seconds", duration.Seconds())

	forkEnvPath := fmt.Sprintf("%s/envs/aqueduct_python%s/lib/python%s/site-packages", e.CondaPath, e.PythonVersion, e.PythonVersion)
	log.Infof("forkEnvPath is %s", forkEnvPath)

	forkArgs := []string{
		"develop",
		"-n",
		e.Name(),
		forkEnvPath,
	}

	start = time.Now()
	_, _, err = lib_utils.RunCmd(CondaCmdPrefix, forkArgs...)
	if err != nil {
		return err
	}
	duration = time.Since(start)
	log.Infof("Conda fork took %d seconds", duration.Seconds())

	log.Info("Printing dependencies...")
	for _, d := range e.Dependencies {
		fmt.Println(d)
	}

	// Then, we use pip3 to install dependencies inside this new Conda env.
	installArgs := append([]string{
		"run",
		"-n",
		e.Name(),
		"pip3",
		"install",
	}, e.Dependencies...)

	start = time.Now()
	_, _, err = lib_utils.RunCmd(CondaCmdPrefix, installArgs...)
	if err != nil {
		return err
	}
	duration = time.Since(start)
	log.Infof("Conda dep install took %d seconds", duration.Seconds())

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
func (e *ExecutionEnvironment) DeleteEnv() error {
	return deleteCondaEnv(e.Name())
}

// GetExecEnvFromDB returns an exec env object from DB by its hash.
// It returns database.ErrNoRows if there is no match.
func GetExecEnvFromDB(
	ctx context.Context,
	hash uuid.UUID,
	execEnvReader db_exec_env.Reader,
	db database.Database,
) (*ExecutionEnvironment, error) {
	dbExecEnv, err := execEnvReader.GetExecutionEnvironmentByHash(ctx, hash, db)
	if err != nil {
		return nil, err
	}

	return newFromDBExecutionEnvironment(dbExecEnv), nil
}

func baseEnvNameByVersion(pythonVersion string) string {
	return fmt.Sprintf("aqueduct_python%s", pythonVersion)
}

// createBaseEnvs creates base environments given the base conda path.
func createBaseEnvs(condaBasePath string) error {
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
			log.Errorf("Failed to delete env %s: %v", env.Id.String(), err)
		}
	}
}

func newFromDBExecutionEnvironment(
	dbExecEnv *db_exec_env.DBExecutionEnvironment,
) *ExecutionEnvironment {
	return &ExecutionEnvironment{
		Id:            dbExecEnv.Id,
		PythonVersion: dbExecEnv.Spec.PythonVersion,
		Dependencies:  dbExecEnv.Spec.Dependencies,
	}
}

func GetExecutionEnvironmentsMapByOperatorIDs(
	ctx context.Context,
	opIDs []uuid.UUID,
	envReader db_exec_env.Reader,
	db database.Database,
) (map[uuid.UUID]ExecutionEnvironment, error) {
	dbEnvMap, err := envReader.GetExecutionEnvironmentsMapByOperatorID(
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

// CreateMissingAndSyncExistingEnvs env %s: %vistingEnvs` keep argerrsync with DB.
// In other words, it creates new db rows for missing envs
// and fetch existing ones.
//
// Returns a map with the original key, mapped to the synced
// env object from the DB rows.
func CreateMissingAndSyncExistingEnvs(
	ctx context.Context,
	envReader db_exec_env.Reader,
	envWriter db_exec_env.Writer,
	envs map[uuid.UUID]ExecutionEnvironment,
	db database.Database,
) (map[uuid.UUID]ExecutionEnvironment, error) {
	// visitedResults is an envHash to boolean mapping
	// to track already visited envHash. This helps reduce
	// the number of DB access.
	visitedResults := make(map[uuid.UUID]ExecutionEnvironment, len(envs))
	addedEnvs := make([]ExecutionEnvironment, 0, len(envs))
	results := make(map[uuid.UUID]ExecutionEnvironment, len(envs))
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
			envReader,
			db,
		)

		// Env is missing
		if err == database.ErrNoRows {
			err = env.CreateDBRecord(ctx, envWriter, db)
			if err != nil {
				deleteEnvs(addedEnvs)
				return nil, err
			}

			err = env.CreateEnv()
			if err != nil {
				deleteEnvs(addedEnvs)
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

	return results, nil
}

func GetUnusedExecutionEnvironmentIDs(
	ctx context.Context,
	envReader db_exec_env.Reader,
	db database.Database,
) ([]uuid.UUID, error) {
	dbEnvs, err := envReader.GetUnusedExecutionEnvironments(
		ctx, db,
	)
	if err != nil {
		return nil, err
	}

	results := make([]uuid.UUID, 0, len(dbEnvs))
	for _, dbEnv := range dbEnvs {
		results = append(results, dbEnv.Id)
	}

	return results, nil
}

// CleanupUnusedEnvironments is asynchronously executed in a Go routine in a best-effort
// fashion, so we log the errors instead of returning them.
func CleanupUnusedEnvironments(
	ctx context.Context,
	envReader db_exec_env.Reader,
	envWriter db_exec_env.Writer,
	db database.Database,
) {
	envIDs, err := GetUnusedExecutionEnvironmentIDs(ctx, envReader, db)
	if err != nil {
		log.Errorf("Error getting unused execution environments: %v", err)
		return
	}

	var errIDs []uuid.UUID
	var deletedIDs []uuid.UUID

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
			errIDs = append(errIDs, envID)
		} else {
			deletedIDs = append(deletedIDs, envID)
		}
	}

	err = envWriter.DeleteExecutionEnvironments(ctx, deletedIDs, db)
	if err != nil {
		log.Errorf("Error deleting database records of unused Conda environments: %v", err)
	}

	if len(errIDs) != 0 {
		log.Errorf("Error garbage collecting Conda environments: %v", errIDs)
	}
}
