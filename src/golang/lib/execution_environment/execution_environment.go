package execution_environment

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"sort"

	"github.com/aqueducthq/aqueduct/lib"
	db_exec_env "github.com/aqueducthq/aqueduct/lib/collections/execution_environment"
	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

const condaCmdPrefix = "conda"

type ExecutionEnvironment struct {
	Id            uuid.UUID `json:"id"`
	PythonVersion string    `json:"python_version"`
	Dependencies  []string  `json:"dependencies"`
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
		db_exec_env.Spec{
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
	sliceToHash := append(e.Dependencies, e.PythonVersion)
	sort.Strings(sliceToHash)

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

	err := runCmd(condaCmdPrefix, createArgs...)
	if err != nil {
		return err
	}

	// Then, we use pip3 to install dependencies inside this new Conda env.
	// We manually add the aqueduct-ml package because the env sent from
	// the client may not contain this required package.
	installArgs := append([]string{
		"run",
		"-n",
		e.Name(),
		"pip3",
		"install",
		fmt.Sprintf("aqueduct-ml==%s", lib.ServerVersionNumber),
	}, e.Dependencies...)

	err = runCmd(condaCmdPrefix, installArgs...)
	if err != nil {
		return err
	}

	return nil
}

// DeleteEnv deletes the Conda environment if it exists.
func (e *ExecutionEnvironment) DeleteEnv() error {
	deleteArgs := []string{
		"env",
		"remove",
		"-n",
		e.Name(),
	}

	return runCmd(condaCmdPrefix, deleteArgs...)
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

	return &ExecutionEnvironment{
		Id:            dbExecEnv.Id,
		PythonVersion: dbExecEnv.Spec.PythonVersion,
		Dependencies:  dbExecEnv.Spec.Dependencies,
	}, nil
}

func IsCondaConnected(
	ctx context.Context,
	userId uuid.UUID,
	integrationReader integration.Reader,
	db database.Database,
) (bool, error) {
	integrations, err := integrationReader.GetIntegrationsByServiceAndUser(
		ctx,
		integration.Conda,
		userId,
		db,
	)
	if err != nil {
		return false, err
	}

	return len(integrations) > 0, nil
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
