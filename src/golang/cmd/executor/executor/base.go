package executor

import (
	"context"

	"github.com/aqueducthq/aqueduct/config"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/repos/sqlite"
	"github.com/aqueducthq/aqueduct/lib/vault"
)

// `BaseExecutor` contains shared attributes for most implementations.
// It does not implement the `Run` method.
type BaseExecutor struct {
	JobManager job.JobManager
	Vault      vault.Vault
	Database   database.Database
	*Repos
}

func NewBaseExecutor(conf *job.ExecutorConfiguration) (*BaseExecutor, error) {
	DB, err := database.NewDatabase(conf.Database)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			DB.Close()
		}
	}()

	schemaVersionRepo := sqlite.NewSchemaVersionRepo()

	if err := requireSchemaVersion(
		context.Background(),
		models.CurrentSchemaVersion,
		schemaVersionRepo,
		DB,
	); err != nil {
		return nil, errors.Wrap(err, "Found incompatible database schema version.")
	}

	jobManager, err := job.NewJobManager(conf.JobManager)
	if err != nil {
		return nil, err
	}

	storageConfig := config.Storage()
	vault, err := vault.NewVault(&storageConfig, config.EncryptionKey())
	if err != nil {
		return nil, err
	}

	return &BaseExecutor{
		JobManager: jobManager,
		Vault:      vault,
		Database:   DB,
		Repos:      createRepos(),
	}, nil
}

func (ex *BaseExecutor) Close() {
	ex.Database.Close()
}

// requireSchemaVersion returns an error if the database schema version is
// not at least version.
func requireSchemaVersion(
	ctx context.Context,
	version int64,
	schemaVersionRepo repos.SchemaVersion,
	DB database.Database,
) error {
	currentVersion, err := schemaVersionRepo.GetCurrent(ctx, DB)
	if err != nil {
		return err
	}

	if currentVersion.Version < version {
		return errors.Newf(
			"Current version is %d, but %d is required.",
			currentVersion.Version,
			version,
		)
	}
	return nil
}
