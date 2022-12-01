package executor

import (
	"github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact_result"
	exec_env "github.com/aqueducthq/aqueduct/lib/collections/execution_environment"
	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	"github.com/aqueducthq/aqueduct/lib/collections/notification"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/collections/operator_result"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_edge"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_result"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_watcher"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/engine"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/repos/sqlite"
)

type Readers struct {
	WorkflowReader             workflow.Reader
	WorkflowDagReader          workflow_dag.Reader
	OperatorReader             operator.Reader
	ArtifactReader             artifact.Reader
	WorkflowDagEdgeReader      workflow_dag_edge.Reader
	IntegrationReader          integration.Reader
	WorkflowDagResultReader    workflow_dag_result.Reader
	OperatorResultReader       operator_result.Reader
	ArtifactResultReader       artifact_result.Reader
	ExecutionEnvironmentReader exec_env.Reader
}

type Writers struct {
	WorkflowWriter          workflow.Writer
	WorkflowDagWriter       workflow_dag.Writer
	WorkflowDagResultWriter workflow_dag_result.Writer
	WorkflowDagEdgeWriter   workflow_dag_edge.Writer
	WorkflowWatcherWriter   workflow_watcher.Writer
	OperatorWriter          operator.Writer
	OperatorResultWriter    operator_result.Writer
	ArtifactWriter          artifact.Writer
	ArtifactResultWriter    artifact_result.Writer
	NotificationWriter      notification.Writer
}

type Repos struct {
	DAGResultRepo repos.DAGResult
	WatcherRepo   repos.Watcher
}

func createReaders(dbConf *database.DatabaseConfig) (*Readers, error) {
	workflowReader, err := workflow.NewReader(dbConf)
	if err != nil {
		return nil, err
	}

	workflowDagReader, err := workflow_dag.NewReader(dbConf)
	if err != nil {
		return nil, err
	}

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

	workflowDagResultReader, err := workflow_dag_result.NewReader(dbConf)
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
		WorkflowReader:             workflowReader,
		WorkflowDagReader:          workflowDagReader,
		OperatorReader:             operatorReader,
		ArtifactReader:             artifactReader,
		WorkflowDagEdgeReader:      workflowDagEdgeReader,
		IntegrationReader:          integrationReader,
		WorkflowDagResultReader:    workflowDagResultReader,
		OperatorResultReader:       operatorResultReader,
		ArtifactResultReader:       artifactResultReader,
		ExecutionEnvironmentReader: execEnvReader,
	}, nil
}

func createWriters(dbConf *database.DatabaseConfig) (*Writers, error) {
	workflowWriter, err := workflow.NewWriter(dbConf)
	if err != nil {
		return nil, err
	}

	workflowDagWriter, err := workflow_dag.NewWriter(dbConf)
	if err != nil {
		return nil, err
	}

	workflowDagResultWriter, err := workflow_dag_result.NewWriter(dbConf)
	if err != nil {
		return nil, err
	}

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
		WorkflowWriter:          workflowWriter,
		WorkflowDagWriter:       workflowDagWriter,
		WorkflowDagResultWriter: workflowDagResultWriter,
		WorkflowDagEdgeWriter:   workflowDagEdgeWriter,
		WorkflowWatcherWriter:   workflowWatcherWriter,
		OperatorWriter:          operatorWriter,
		OperatorResultWriter:    operatorResultWriter,
		ArtifactWriter:          artifactWriter,
		ArtifactResultWriter:    artifactResultWriter,
		NotificationWriter:      notificationWriter,
	}, nil
}

func createRepos() *Repos {
	return &Repos{
		DAGResultRepo: sqlite.NewDAGResultRepo(),
		WatcherRepo:   sqlite.NewWatcherRepo(),
	}
}

func getEngineReaders(readers *Readers) *engine.EngineReaders {
	return &engine.EngineReaders{
		WorkflowReader:             readers.WorkflowReader,
		WorkflowDagReader:          readers.WorkflowDagReader,
		WorkflowDagEdgeReader:      readers.WorkflowDagEdgeReader,
		WorkflowDagResultReader:    readers.WorkflowDagResultReader,
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
		WorkflowWriter:          writers.WorkflowWriter,
		WorkflowDagWriter:       writers.WorkflowDagWriter,
		WorkflowDagEdgeWriter:   writers.WorkflowDagEdgeWriter,
		WorkflowDagResultWriter: writers.WorkflowDagResultWriter,
		OperatorWriter:          writers.OperatorWriter,
		OperatorResultWriter:    writers.OperatorResultWriter,
		ArtifactWriter:          writers.ArtifactWriter,
		ArtifactResultWriter:    writers.ArtifactResultWriter,
		NotificationWriter:      writers.NotificationWriter,
	}
}

func getEngineRepos(repos *Repos) *engine.Repos {
	return &engine.Repos{
		DAGResultRepo: repos.DAGResultRepo,
		WatcherRepo:   repos.WatcherRepo,
	}
}
