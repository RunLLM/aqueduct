package tests

import (
	"context"
	"os"

	"github.com/aqueducthq/aqueduct/cmd/migrator/migrator"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/repos/sqlite"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

// TestSuite contains all fields that are needed by the database integration
// tests. Each test case should be implemented as a method of TestSuite.
type TestSuite struct {
	suite.Suite
	ctx context.Context

	// List of all repos
	artifact       repos.Artifact
	artifactResult repos.ArtifactResult
	dag            repos.DAG
	dagEdge        repos.DAGEdge
	dagResult      repos.DAGResult
	executionEnvironment       repos.ExecutionEnvironment
	integration    repos.Integration
	notification   repos.Notification
	operator       repos.Operator
	user           repos.User
	watcher        repos.Watcher
	workflow       repos.Workflow

	DB database.Database
}

// SetupSuite is run only once before all tests. It initializes the database
// connection and creates the repos. It initializes the database schema
// to the latest version.
func (ts *TestSuite) SetupSuite() {
	DB, err := database.NewSqliteInMemoryDatabase(&database.SqliteConfig{})
	if err != nil {
		ts.T().Errorf("Unable to create SQLite client: %v", err)
	}

	ts.ctx = context.Background()
	ts.DB = DB

	// Initialize repos
	ts.artifact = sqlite.NewArtifactRepo()
	ts.artifactResult = sqlite.NewArtifactResultRepo()
	ts.dag = sqlite.NewDAGRepo()
	ts.dagEdge = sqlite.NewDAGEdgeRepo()
	ts.dagResult = sqlite.NewDAGResultRepo()
	ts.executionEnvironment = sqlite.NewExecutionEnvironmentRepo()
	ts.integration = sqlite.NewIntegrationRepo()
	ts.notification = sqlite.NewNotificationRepo()
	ts.operator = sqlite.NewOperatorRepo()
	ts.user = sqlite.NewUserRepo()
	ts.watcher = sqlite.NewWatcherRepo()
	ts.workflow = sqlite.NewWorklowRepo()

	// Init database schema
	if err := initDBSchema(DB); err != nil {
		DB.Close()
		ts.T().Errorf("Unable to initialize database schema: %v", err)
	}
}

func initDBSchema(DB database.Database) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// The schema change logic must be invoked from the `golang/` directory
	if err := os.Chdir("../../.."); err != nil {
		return err
	}

	defer func() {
		chdirErr := os.Chdir(cwd)
		if chdirErr != nil {
			log.Errorf("Error when changing cwd: %v", chdirErr)
		}
	}()

	if err := migrator.GoTo(context.Background(), models.SchemaVersion, DB); err != nil {
		return err
	}

	return nil
}

// TearDownTest is run after each test finishes.
func (ts *TestSuite) TearDownTest() {
	// Clear all of the tables
	query := `
	DELETE FROM app_user;
	DELETE FROM execution_environment;
	DELETE FROM integration;
	DELETE FROM workflow;
	DELETE FROM workflow_dag;
	DELETE FROM workflow_dag_result;
	DELETE FROM workflow_dag_edge;
	DELETE FROM operator;
	DELETE FROM operator_result;
	DELETE FROM artifact;
	DELETE FROM artifact_result;
	DELETE FROM notification;
	;
	`
	if err := ts.DB.Execute(ts.ctx, query); err != nil {
		ts.T().Errorf("Unable to clear database: %v", err)
	}
}

// TearDownSuite is run after all tests complete.
func (ts *TestSuite) TearDownSuite() {
	ts.DB.Close()
}
