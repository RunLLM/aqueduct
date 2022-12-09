package server

import (
	"github.com/aqueducthq/aqueduct/cmd/server/queries"
	exec_env "github.com/aqueducthq/aqueduct/lib/collections/execution_environment"
	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/collections/schema_version"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/engine"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/repos/sqlite"
)

type Repos struct {
	ArtifactRepo       repos.Artifact
	ArtifactResultRepo repos.ArtifactResult
	DAGRepo            repos.DAG
	DAGEdgeRepo        repos.DAGEdge
	DAGResultRepo      repos.DAGResult
	NotificationRepo   repos.Notification
	OperatorRepo       repos.Operator
	OperatorResultRepo repos.OperatorResult
	UserRepo           repos.User
	WatcherRepo        repos.Watcher
	WorkflowRepo       repos.Workflow
}

type Readers struct {
	IntegrationReader          integration.Reader
	OperatorReader             operator.Reader
	WorkflowReader             workflow.Reader
	SchemaVersionReader        schema_version.Reader
	CustomReader               queries.Reader
	ExecutionEnvironmentReader exec_env.Reader
}

type Writers struct {
	IntegrationWriter          integration.Writer
	ExecutionEnvironmentWriter exec_env.Writer
}

func CreateRepos() *Repos {
	return &Repos{
		ArtifactRepo:       sqlite.NewArtifactRepo(),
		ArtifactResultRepo: sqlite.NewArtifactResultRepo(),
		DAGRepo:            sqlite.NewDAGRepo(),
		DAGEdgeRepo:        sqlite.NewDAGEdgeRepo(),
		DAGResultRepo:      sqlite.NewDAGResultRepo(),
		NotificationRepo:   sqlite.NewNotificationRepo(),
		OperatorRepo:       sqlite.NewOperatorRepo(),
		OperatorResultRepo: sqlite.NewOperatorResultRepo(),
		UserRepo:           sqlite.NewUserRepo(),
		WatcherRepo:        sqlite.NewWatcherRepo(),
		WorkflowRepo:       sqlite.NewWorklowRepo(),
	}
}

func CreateReaders(dbConfig *database.DatabaseConfig) (*Readers, error) {
	integrationReader, err := integration.NewReader(dbConfig)
	if err != nil {
		return nil, err
	}

	operatorReader, err := operator.NewReader(dbConfig)
	if err != nil {
		return nil, err
	}

	workflowReader, err := workflow.NewReader(dbConfig)
	if err != nil {
		return nil, err
	}

	schemaVersionReader, err := schema_version.NewReader(dbConfig)
	if err != nil {
		return nil, err
	}

	queriesReader, err := queries.NewReader(dbConfig)
	if err != nil {
		return nil, err
	}

	execEnvReader, err := exec_env.NewReader(dbConfig)
	if err != nil {
		return nil, err
	}

	return &Readers{
		IntegrationReader:          integrationReader,
		OperatorReader:             operatorReader,
		WorkflowReader:             workflowReader,
		SchemaVersionReader:        schemaVersionReader,
		CustomReader:               queriesReader,
		ExecutionEnvironmentReader: execEnvReader,
	}, nil
}

func CreateWriters(dbConfig *database.DatabaseConfig) (*Writers, error) {
	integrationWriter, err := integration.NewWriter(dbConfig)
	if err != nil {
		return nil, err
	}

	execEnvWriter, err := exec_env.NewWriter(dbConfig)
	if err != nil {
		return nil, err
	}

	return &Writers{
		IntegrationWriter:          integrationWriter,
		ExecutionEnvironmentWriter: execEnvWriter,
	}, nil
}

func GetEngineReaders(readers *Readers) *engine.EngineReaders {
	return &engine.EngineReaders{
		OperatorReader:             readers.OperatorReader,
		IntegrationReader:          readers.IntegrationReader,
		ExecutionEnvironmentReader: readers.ExecutionEnvironmentReader,
	}
}

func GetEngineRepos(repos *Repos) *engine.Repos {
	return &engine.Repos{
		ArtifactRepo:       repos.ArtifactRepo,
		ArtifactResultRepo: repos.ArtifactResultRepo,
		DAGRepo:            repos.DAGRepo,
		DAGEdgeRepo:        repos.DAGEdgeRepo,
		DAGResultRepo:      repos.DAGResultRepo,
		NotificationRepo:   repos.NotificationRepo,
		OperatorRepo:       repos.OperatorRepo,
		OperatorResultRepo: repos.OperatorResultRepo,
		WatcherRepo:        repos.WatcherRepo,
		WorkflowRepo:       repos.WorkflowRepo,
	}
}
