package executor

import (
	"github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact_result"
	exec_env "github.com/aqueducthq/aqueduct/lib/collections/execution_environment"
	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	"github.com/aqueducthq/aqueduct/lib/collections/notification"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/collections/operator_result"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_edge"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_watcher"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/engine"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/repos/sqlite"
)

type Readers struct {
	OperatorReader             operator.Reader
	ArtifactReader             artifact.Reader
	WorkflowDagEdgeReader      workflow_dag_edge.Reader
	IntegrationReader          integration.Reader
	OperatorResultReader       operator_result.Reader
	ArtifactResultReader       artifact_result.Reader
	ExecutionEnvironmentReader exec_env.Reader
}

type Writers struct {
	WorkflowDagEdgeWriter workflow_dag_edge.Writer
	WorkflowWatcherWriter workflow_watcher.Writer
	OperatorWriter        operator.Writer
	OperatorResultWriter  operator_result.Writer
	ArtifactWriter        artifact.Writer
	ArtifactResultWriter  artifact_result.Writer
	NotificationWriter    notification.Writer
}

type Repos struct {
	DAGRepo       repos.DAG
	DAGResultRepo repos.DAGResult
	WatcherRepo   repos.Watcher
	WorkflowRepo  repos.Workflow
}

func createReaders(dbConf *database.DatabaseConfig) (*Readers, error) {
	operatorReader, err := operator.NewReader(dbConf)
	if err != nil {
		return nil, err
	}

	artifactReader, err := artifact.NewReader(dbConf)
	if err != nil {
		return nil, err
	}

	workflowDagEdgeReader, err := workflow_dag_edge.NewReader(dbConf)
	if err != nil {
		return nil, err
	}

	integrationReader, err := integration.NewReader(dbConf)
	if err != nil {
		return nil, err
	}

	operatorResultReader, err := operator_result.NewReader(dbConf)
	if err != nil {
		return nil, err
	}

	artifactResultReader, err := artifact_result.NewReader(dbConf)
	if err != nil {
		return nil, err
	}

	execEnvReader, err := exec_env.NewReader(dbConf)
	if err != nil {
		return nil, err
	}

	return &Readers{
		OperatorReader:             operatorReader,
		ArtifactReader:             artifactReader,
		WorkflowDagEdgeReader:      workflowDagEdgeReader,
		IntegrationReader:          integrationReader,
		OperatorResultReader:       operatorResultReader,
		ArtifactResultReader:       artifactResultReader,
		ExecutionEnvironmentReader: execEnvReader,
	}, nil
}

func createWriters(dbConf *database.DatabaseConfig) (*Writers, error) {
	workflowDagEdgeWriter, err := workflow_dag_edge.NewWriter(dbConf)
	if err != nil {
		return nil, err
	}

	workflowWatcherWriter, err := workflow_watcher.NewWriter(dbConf)
	if err != nil {
		return nil, err
	}

	operatorWriter, err := operator.NewWriter(dbConf)
	if err != nil {
		return nil, err
	}

	operatorResultWriter, err := operator_result.NewWriter(dbConf)
	if err != nil {
		return nil, err
	}

	artifactWriter, err := artifact.NewWriter(dbConf)
	if err != nil {
		return nil, err
	}

	artifactResultWriter, err := artifact_result.NewWriter(dbConf)
	if err != nil {
		return nil, err
	}

	notificationWriter, err := notification.NewWriter(dbConf)
	if err != nil {
		return nil, err
	}

	return &Writers{
		WorkflowDagEdgeWriter: workflowDagEdgeWriter,
		WorkflowWatcherWriter: workflowWatcherWriter,
		OperatorWriter:        operatorWriter,
		OperatorResultWriter:  operatorResultWriter,
		ArtifactWriter:        artifactWriter,
		ArtifactResultWriter:  artifactResultWriter,
		NotificationWriter:    notificationWriter,
	}, nil
}

func createRepos() *Repos {
	return &Repos{
		DAGRepo:       sqlite.NewDAGRepo(),
		DAGResultRepo: sqlite.NewDAGResultRepo(),
		WatcherRepo:   sqlite.NewWatcherRepo(),
		WorkflowRepo:  sqlite.NewWorklowRepo(),
	}
}

func getEngineReaders(readers *Readers) *engine.EngineReaders {
	return &engine.EngineReaders{
		WorkflowDagEdgeReader:      readers.WorkflowDagEdgeReader,
		OperatorReader:             readers.OperatorReader,
		OperatorResultReader:       readers.OperatorResultReader,
		ArtifactReader:             readers.ArtifactReader,
		ArtifactResultReader:       readers.ArtifactResultReader,
		IntegrationReader:          readers.IntegrationReader,
		ExecutionEnvironmentReader: readers.ExecutionEnvironmentReader,
	}
}

func getEngineWriters(writers *Writers) *engine.EngineWriters {
	return &engine.EngineWriters{
		WorkflowDagEdgeWriter: writers.WorkflowDagEdgeWriter,
		OperatorWriter:        writers.OperatorWriter,
		OperatorResultWriter:  writers.OperatorResultWriter,
		ArtifactWriter:        writers.ArtifactWriter,
		ArtifactResultWriter:  writers.ArtifactResultWriter,
		NotificationWriter:    writers.NotificationWriter,
	}
}

func getEngineRepos(repos *Repos) *engine.Repos {
	return &engine.Repos{
		DAGRepo:       repos.DAGRepo,
		DAGResultRepo: repos.DAGResultRepo,
		WatcherRepo:   repos.WatcherRepo,
		WorkflowRepo:  repos.WorkflowRepo,
	}
}
