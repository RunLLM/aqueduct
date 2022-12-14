package server

import (
	"github.com/aqueducthq/aqueduct/lib/collections/schema_version"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/engine"
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
	UserRepo                 repos.User
	WatcherRepo              repos.Watcher
	WorkflowRepo             repos.Workflow
}

type Readers struct {
	SchemaVersionReader schema_version.Reader
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
		UserRepo:                 sqlite.NewUserRepo(),
		WatcherRepo:              sqlite.NewWatcherRepo(),
		WorkflowRepo:             sqlite.NewWorklowRepo(),
	}
}

func CreateReaders(dbConfig *database.DatabaseConfig) (*Readers, error) {
	schemaVersionReader, err := schema_version.NewReader(dbConfig)
	if err != nil {
		return nil, err
	}

	return &Readers{
		SchemaVersionReader: schemaVersionReader,
	}, nil
}

func GetEngineRepos(repos *Repos) *engine.Repos {
	return &engine.Repos{
		ArtifactRepo:             repos.ArtifactRepo,
		ArtifactResultRepo:       repos.ArtifactResultRepo,
		DAGRepo:                  repos.DAGRepo,
		DAGEdgeRepo:              repos.DAGEdgeRepo,
		DAGResultRepo:            repos.DAGResultRepo,
		ExecutionEnvironmentRepo: repos.ExecutionEnvironmentRepo,
		NotificationRepo:         repos.NotificationRepo,
		OperatorRepo:             repos.OperatorRepo,
		OperatorResultRepo:       repos.OperatorResultRepo,
		WatcherRepo:              repos.WatcherRepo,
		WorkflowRepo:             repos.WorkflowRepo,
	}
}
