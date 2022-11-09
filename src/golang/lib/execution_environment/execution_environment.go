package execution_environment

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"sort"

	"github.com/aqueducthq/aqueduct/lib"
	dbExecEnv "github.com/aqueducthq/aqueduct/lib/collections/execution_environment"
	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/google/uuid"
)

const condaCmdPrefix = "conda"

type ExecutionEnvironment struct {
	Id            uuid.UUID `json:"id"`
	PythonVersion string    `json:"python_version"`
	Dependencies  []string  `json:"dependencies"`

	execEnvWriter dbExecEnv.Writer
	db            database.Database
}

func (e *ExecutionEnvironment) CreateDBRecord(ctx context.Context) error {
	hash, err := e.Hash()
	if err != nil {
		return err
	}

	_, err = e.execEnvWriter.CreateExecutionEnvironment(
		ctx,
		dbExecEnv.Spec{
			PythonVersion: e.PythonVersion,
			Dependencies:  e.Dependencies,
		},
		hash,
		e.db,
	)
	return err
}

func (e *ExecutionEnvironment) DeleteDBRecord(ctx context.Context) error {
	return e.execEnvWriter.DeleteExecutionEnvironment(ctx, e.Id, e.db)
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
	execEnvReader dbExecEnv.Reader,
	execEnvWriter dbExecEnv.Writer,
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
		execEnvWriter: execEnvWriter,
		db:            db,
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
