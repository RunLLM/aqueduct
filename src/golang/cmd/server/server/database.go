package server

import (
	"github.com/aqueducthq/aqueduct/internal/server/queries"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact_result"
	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	"github.com/aqueducthq/aqueduct/lib/collections/notification"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/collections/operator_result"
	"github.com/aqueducthq/aqueduct/lib/collections/schema_version"
	"github.com/aqueducthq/aqueduct/lib/collections/user"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_edge"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_result"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_watcher"
	"github.com/aqueducthq/aqueduct/lib/database"
)

type Readers struct {
	UserReader              user.Reader
	IntegrationReader       integration.Reader
	NotificationReader      notification.Reader
	ArtifactReader          artifact.Reader
	ArtifactResultReader    artifact_result.Reader
	OperatorReader          operator.Reader
	OperatorResultReader    operator_result.Reader
	WorkflowReader          workflow.Reader
	WorkflowDagReader       workflow_dag.Reader
	WorkflowDagEdgeReader   workflow_dag_edge.Reader
	WorkflowWatcherReader   workflow_watcher.Reader
	WorkflowDagResultReader workflow_dag_result.Reader
	SchemaVersionReader     schema_version.Reader
	CustomReader            queries.Reader
}

type Writers struct {
	UserWriter              user.Writer
	IntegrationWriter       integration.Writer
	NotificationWriter      notification.Writer
	ArtifactWriter          artifact.Writer
	ArtifactResultWriter    artifact_result.Writer
	OperatorWriter          operator.Writer
	OperatorResultWriter    operator_result.Writer
	WorkflowWriter          workflow.Writer
	WorkflowDagWriter       workflow_dag.Writer
	WorkflowDagEdgeWriter   workflow_dag_edge.Writer
	WorkflowWatcherWriter   workflow_watcher.Writer
	WorkflowDagResultWriter workflow_dag_result.Writer
}

func CreateReaders(dbConfig *database.DatabaseConfig) (*Readers, error) {
	userReader, err := user.NewReader(dbConfig)
	if err != nil {
		return nil, err
	}

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

	workflowDagReader, err := workflow_dag.NewReader(dbConfig)
	if err != nil {
		return nil, err
	}

	workflowDagEdgeReader, err := workflow_dag_edge.NewReader(dbConfig)
	if err != nil {
		return nil, err
	}

	workflowWatcherReader, err := workflow_watcher.NewReader(dbConfig)
	if err != nil {
		return nil, err
	}

	workflowDagResultReader, err := workflow_dag_result.NewReader(dbConfig)
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

	return &Readers{
		UserReader:              userReader,
		IntegrationReader:       integrationReader,
		NotificationReader:      notificationReader,
		ArtifactReader:          artifactReader,
		ArtifactResultReader:    artifactResultReader,
		OperatorReader:          operatorReader,
		OperatorResultReader:    operatorResultReader,
		WorkflowReader:          workflowReader,
		WorkflowDagReader:       workflowDagReader,
		WorkflowDagEdgeReader:   workflowDagEdgeReader,
		WorkflowWatcherReader:   workflowWatcherReader,
		WorkflowDagResultReader: workflowDagResultReader,
		SchemaVersionReader:     schemaVersionReader,
		CustomReader:            queriesReader,
	}, nil
}

func CreateWriters(dbConfig *database.DatabaseConfig) (*Writers, error) {
	userWriter, err := user.NewWriter(dbConfig)
	if err != nil {
		return nil, err
	}

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

	workflowWriter, err := workflow.NewWriter(dbConfig)
	if err != nil {
		return nil, err
	}

	workflowDagWriter, err := workflow_dag.NewWriter(dbConfig)
	if err != nil {
		return nil, err
	}

	workflowDagEdgeWriter, err := workflow_dag_edge.NewWriter(dbConfig)
	if err != nil {
		return nil, err
	}

	workflowWatcherWriter, err := workflow_watcher.NewWriter(dbConfig)
	if err != nil {
		return nil, err
	}

	workflowDagResultWriter, err := workflow_dag_result.NewWriter(dbConfig)
	if err != nil {
		return nil, err
	}

	return &Writers{
		UserWriter:              userWriter,
		IntegrationWriter:       integrationWriter,
		NotificationWriter:      notificationWriter,
		ArtifactWriter:          artifactWriter,
		ArtifactResultWriter:    artifactResultWriter,
		OperatorWriter:          operatorWriter,
		OperatorResultWriter:    operatorResultWriter,
		WorkflowWriter:          workflowWriter,
		WorkflowDagWriter:       workflowDagWriter,
		WorkflowDagEdgeWriter:   workflowDagEdgeWriter,
		WorkflowWatcherWriter:   workflowWatcherWriter,
		WorkflowDagResultWriter: workflowDagResultWriter,
	}, nil
}
