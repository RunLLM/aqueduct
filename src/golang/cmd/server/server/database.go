package server

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/engine"
	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/repos/sqlite"
)

type Repos struct {
	ArtifactRepo             repos.Artifact
	ArtifactResultRepo       repos.ArtifactResult
	DAGRepo                  repos.DAG
	DAGEdgeRepo              repos.DAGEdge
	DAGResultRepo            repos.DAGResult
	ExecutionEnvironmentRepo repos.ExecutionEnvironment
	IntegrationRepo          repos.Integration
	NotificationRepo         repos.Notification
	OperatorRepo             repos.Operator
	OperatorResultRepo       repos.OperatorResult
	SchemaVersionRepo        repos.SchemaVersion
	UserRepo                 repos.User
	WatcherRepo              repos.Watcher
	WorkflowRepo             repos.Workflow
}

func CreateRepos() *Repos {
	return &Repos{
		ArtifactRepo:             sqlite.NewArtifactRepo(),
		ArtifactResultRepo:       sqlite.NewArtifactResultRepo(),
		DAGRepo:                  sqlite.NewDAGRepo(),
		DAGEdgeRepo:              sqlite.NewDAGEdgeRepo(),
		DAGResultRepo:            sqlite.NewDAGResultRepo(),
		ExecutionEnvironmentRepo: sqlite.NewExecutionEnvironmentRepo(),
		IntegrationRepo:          sqlite.NewIntegrationRepo(),
		NotificationRepo:         sqlite.NewNotificationRepo(),
		OperatorRepo:             sqlite.NewOperatorRepo(),
		OperatorResultRepo:       sqlite.NewOperatorResultRepo(),
		SchemaVersionRepo:        sqlite.NewSchemaVersionRepo(),
		UserRepo:                 sqlite.NewUserRepo(),
		WatcherRepo:              sqlite.NewWatcherRepo(),
		WorkflowRepo:             sqlite.NewWorklowRepo(),
	}
}

func GetEngineRepos(repos *Repos) *engine.Repos {
	return &engine.Repos{
		ArtifactRepo:             repos.ArtifactRepo,
		ArtifactResultRepo:       repos.ArtifactResultRepo,
		DAGRepo:                  repos.DAGRepo,
		DAGEdgeRepo:              repos.DAGEdgeRepo,
		DAGResultRepo:            repos.DAGResultRepo,
		ExecutionEnvironmentRepo: repos.ExecutionEnvironmentRepo,
		IntegrationRepo:          repos.IntegrationRepo,
		NotificationRepo:         repos.NotificationRepo,
		OperatorRepo:             repos.OperatorRepo,
		OperatorResultRepo:       repos.OperatorResultRepo,
		WatcherRepo:              repos.WatcherRepo,
		WorkflowRepo:             repos.WorkflowRepo,
	}
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
