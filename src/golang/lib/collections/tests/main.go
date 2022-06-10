package tests

import (
	"github.com/aqueducthq/aqueduct/cmd/server/queries"
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

var (
	db      database.Database
	readers *dbReaders
	writers *dbWriters
)

type dbReaders struct {
	artifactReader          artifact.Reader
	artifactResultReader    artifact_result.Reader
	integrationReader       integration.Reader
	notificationReader      notification.Reader
	operatorReader          operator.Reader
	operatorResultReader    operator_result.Reader
	schemaVersionReader     schema_version.Reader
	userReader              user.Reader
	workflowReader          workflow.Reader
	workflowDagReader       workflow_dag.Reader
	workflowDagEdgeReader   workflow_dag_edge.Reader
	workflowDagResultReader workflow_dag_result.Reader
	workflowWatcherReader   workflow_watcher.Reader
	serverReader            queries.Reader
}

type dbWriters struct {
	artifactWriter          artifact.Writer
	artifactResultWriter    artifact_result.Writer
	integrationWriter       integration.Writer
	notificationWriter      notification.Writer
	operatorWriter          operator.Writer
	operatorResultWriter    operator_result.Writer
	schemaVersionWriter     schema_version.Writer
	userWriter              user.Writer
	workflowWriter          workflow.Writer
	workflowDagWriter       workflow_dag.Writer
	workflowDagEdgeWriter   workflow_dag_edge.Writer
	workflowDagResultWriter workflow_dag_result.Writer
	workflowWatcherWriter   workflow_watcher.Writer
}

func createReaders(dbConfig *database.DatabaseConfig) (*dbReaders, error) {
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

	return &dbReaders{
		userReader:              userReader,
		integrationReader:       integrationReader,
		notificationReader:      notificationReader,
		artifactReader:          artifactReader,
		artifactResultReader:    artifactResultReader,
		operatorReader:          operatorReader,
		operatorResultReader:    operatorResultReader,
		workflowReader:          workflowReader,
		workflowDagReader:       workflowDagReader,
		workflowDagEdgeReader:   workflowDagEdgeReader,
		workflowWatcherReader:   workflowWatcherReader,
		workflowDagResultReader: workflowDagResultReader,
		schemaVersionReader:     schemaVersionReader,
		serverReader:            queriesReader,
	}, nil
}

func createWriters(dbConfig *database.DatabaseConfig) (*dbWriters, error) {
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

	schemaVersionWriter, err := schema_version.NewWriter(dbConfig)
	if err != nil {
		return nil, err
	}

	return &dbWriters{
		userWriter:              userWriter,
		integrationWriter:       integrationWriter,
		notificationWriter:      notificationWriter,
		artifactWriter:          artifactWriter,
		artifactResultWriter:    artifactResultWriter,
		operatorWriter:          operatorWriter,
		operatorResultWriter:    operatorResultWriter,
		workflowWriter:          workflowWriter,
		workflowDagWriter:       workflowDagWriter,
		workflowDagEdgeWriter:   workflowDagEdgeWriter,
		workflowWatcherWriter:   workflowWatcherWriter,
		workflowDagResultWriter: workflowDagResultWriter,
		schemaVersionWriter:     schemaVersionWriter,
	}, nil
}
