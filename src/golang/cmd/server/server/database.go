package server

import (
	"github.com/aqueducthq/aqueduct/cmd/server/queries"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact_result"
	exec_env "github.com/aqueducthq/aqueduct/lib/collections/execution_environment"
	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	"github.com/aqueducthq/aqueduct/lib/collections/notification"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/collections/operator_result"
	"github.com/aqueducthq/aqueduct/lib/collections/schema_version"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/engine"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/repos/sqlite"
)

type Repos struct {
	DAGRepo       repos.DAG
	DAGEdgeRepo   repos.DAGEdge
	DAGResultRepo repos.DAGResult
	UserRepo      repos.User
	WatcherRepo   repos.Watcher
	WorkflowRepo  repos.Workflow
}

type Readers struct {
	IntegrationReader          integration.Reader
	NotificationReader         notification.Reader
	ArtifactReader             artifact.Reader
	ArtifactResultReader       artifact_result.Reader
	OperatorReader             operator.Reader
	OperatorResultReader       operator_result.Reader
	WorkflowReader             workflow.Reader
	SchemaVersionReader        schema_version.Reader
	CustomReader               queries.Reader
	ExecutionEnvironmentReader exec_env.Reader
}

type Writers struct {
	IntegrationWriter          integration.Writer
	NotificationWriter         notification.Writer
	ArtifactWriter             artifact.Writer
	ArtifactResultWriter       artifact_result.Writer
	OperatorWriter             operator.Writer
	OperatorResultWriter       operator_result.Writer
	ExecutionEnvironmentWriter exec_env.Writer
}

func CreateRepos() *Repos {
	return &Repos{
		DAGRepo:       sqlite.NewDAGRepo(),
		DAGEdgeRepo:   sqlite.NewDAGEdgeRepo(),
		DAGResultRepo: sqlite.NewDAGResultRepo(),
		UserRepo:      sqlite.NewUserRepo(),
		WatcherRepo:   sqlite.NewWatcherRepo(),
		WorkflowRepo:  sqlite.NewWorklowRepo(),
	}
}

func CreateReaders(dbConfig *database.DatabaseConfig) (*Readers, error) {
	integrationReader, err := integration.NewReader(dbConfig)
	if err != nil {
		return nil, err
	}

	notificationReader, err := notification.NewReader(dbConfig)
	if err != nil {
		return nil, err
	}

	artifactReader, err := artifact.NewReader(dbConfig)
	if err != nil {
		return nil, err
	}

	artifactResultReader, err := artifact_result.NewReader(dbConfig)
	if err != nil {
		return nil, err
	}

	operatorReader, err := operator.NewReader(dbConfig)
	if err != nil {
		return nil, err
	}

	operatorResultReader, err := operator_result.NewReader(dbConfig)
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
		NotificationReader:         notificationReader,
		ArtifactReader:             artifactReader,
		ArtifactResultReader:       artifactResultReader,
		OperatorReader:             operatorReader,
		OperatorResultReader:       operatorResultReader,
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

	notificationWriter, err := notification.NewWriter(dbConfig)
	if err != nil {
		return nil, err
	}

	artifactWriter, err := artifact.NewWriter(dbConfig)
	if err != nil {
		return nil, err
	}

	artifactResultWriter, err := artifact_result.NewWriter(dbConfig)
	if err != nil {
		return nil, err
	}

	operatorWriter, err := operator.NewWriter(dbConfig)
	if err != nil {
		return nil, err
	}

	operatorResultWriter, err := operator_result.NewWriter(dbConfig)
	if err != nil {
		return nil, err
	}

	execEnvWriter, err := exec_env.NewWriter(dbConfig)
	if err != nil {
		return nil, err
	}

	return &Writers{
		IntegrationWriter:          integrationWriter,
		NotificationWriter:         notificationWriter,
		ArtifactWriter:             artifactWriter,
		ArtifactResultWriter:       artifactResultWriter,
		OperatorWriter:             operatorWriter,
		OperatorResultWriter:       operatorResultWriter,
		ExecutionEnvironmentWriter: execEnvWriter,
	}, nil
}

func GetEngineReaders(readers *Readers) *engine.EngineReaders {
	return &engine.EngineReaders{
		OperatorReader:             readers.OperatorReader,
		OperatorResultReader:       readers.OperatorResultReader,
		ArtifactReader:             readers.ArtifactReader,
		ArtifactResultReader:       readers.ArtifactResultReader,
		IntegrationReader:          readers.IntegrationReader,
		ExecutionEnvironmentReader: readers.ExecutionEnvironmentReader,
	}
}

func GetEngineWriters(writers *Writers) *engine.EngineWriters {
	return &engine.EngineWriters{
		OperatorWriter:       writers.OperatorWriter,
		OperatorResultWriter: writers.OperatorResultWriter,
		ArtifactWriter:       writers.ArtifactWriter,
		ArtifactResultWriter: writers.ArtifactResultWriter,
		NotificationWriter:   writers.NotificationWriter,
	}
}

func GetEngineRepos(repos *Repos) *engine.Repos {
	return &engine.Repos{
		DAGRepo:       repos.DAGRepo,
		DAGEdgeRepo:   repos.DAGEdgeRepo,
		DAGResultRepo: repos.DAGResultRepo,
		WatcherRepo:   repos.WatcherRepo,
		WorkflowRepo:  repos.WorkflowRepo,
	}
}
