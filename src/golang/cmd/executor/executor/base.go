package executor

import (
	"context"

	"github.com/aqueducthq/aqueduct/config"
	"github.com/aqueducthq/aqueduct/lib/collections"
	"github.com/aqueducthq/aqueduct/lib/collections/schema_version"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/dropbox/godropbox/errors"
)

// `BaseExecutor` contains shared attributes for most implementations.
// It does not implement the `Run` method.
type BaseExecutor struct {
	JobManager job.JobManager
	Vault      vault.Vault
	Database   database.Database
	*Readers
	*Writers
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

	schemaVersionReader, err := schema_version.NewReader(db.Config())
	if err != nil {
		return nil, err
	}

	if err := collections.RequireSchemaVersion(
		context.Background(),
		models.SchemaVersion,
		schemaVersionReader,
		db,
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

	readers, err := createReaders(db.Config())
	if err != nil {
		return nil, err
	}

	writers, err := createWriters(db.Config())
	if err != nil {
		return nil, err
	}

	return &BaseExecutor{
		JobManager: jobManager,
		Vault:      vault,
		Database:   db,
		Readers:    readers,
		Writers:    writers,
		Repos:      createRepos(),
	}, nil
}

func (ex *BaseExecutor) Close() {
	ex.Database.Close()
}
