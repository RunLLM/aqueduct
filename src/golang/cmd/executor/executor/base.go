package executor

import (
	"context"

	"github.com/aqueducthq/aqueduct/config"
	"github.com/aqueducthq/aqueduct/lib/collections"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/repos/sqlite"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/dropbox/godropbox/errors"
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
	db, err := database.NewDatabase(conf.Database)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			db.Close()
		}
	}()

	schemaVersionRepo := sqlite.NewSchemaVersionRepo()

	if err := collections.RequireSchemaVersion(
		context.Background(),
		models.CurrentSchemaVersion,
		schemaVersionRepo,
		db,
	); err != nil {
		return nil, errors.Wrap(err, "Found incompatible database schema version.")
	}

	jobManager, err := job.NewJobManager(conf.JobManager)
	if err != nil {
		return nil, err
	}

	vault, err := vault.NewVault(config.Storage(), config.EncryptionKey())
	if err != nil {
		return nil, err
	}

	return &BaseExecutor{
		JobManager: jobManager,
		Vault:      vault,
		Database:   db,
		Repos:      createRepos(),
	}, nil
}

func (ex *BaseExecutor) Close() {
	ex.Database.Close()
}
