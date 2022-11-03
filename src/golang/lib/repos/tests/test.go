package tests

import (
	"context"
	"flag"
	"os"
	"testing"

	"github.com/aqueducthq/aqueduct/cmd/migrator/migrator"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/repos/sqlite"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

var runTests = flag.Bool("database", false, "If this flag is set, the database integration tests will be run.")

type TestSuite struct {
	suite.Suite
	ctx context.Context

	// List of all repos
	workflow repos.Workflow

	db database.Database
}

// SetupSuite is run only once before all tests. It initializes the database
// connection and creates the repos. It initializes the database schema
// to the latest version.
func (ts *TestSuite) SetupSuite() {
	db, err := database.NewSqliteInMemoryDatabase(&database.SqliteConfig{})
	if err != nil {
		ts.T().Errorf("Unable to create SQLite client: %v", err)
	}

	ts.ctx = context.Background()
	ts.db = db

	// Initialize repos
	ts.workflow = sqlite.NewWorklowRepo()

	// Init database schema
	if err := initDBSchema(db); err != nil {
		db.Close()
		ts.T().Errorf("Unable to initialize database schema: %v", err)
	}
}

func initDBSchema(db database.Database) error {
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

	if err := migrator.GoTo(context.Background(), models.SchemaVersion, db); err != nil {
		return err
	}

	return nil
}

// TearDownTest is run after each test finishes.
func (ts *TestSuite) TearDownTest() {
	// Clear all of the tables
	query := `
	DELETE FROM app_user;
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
	if err := ts.db.Execute(ts.ctx, query); err != nil {
		ts.T().Errorf("Unable to clear database: %v", err)
	}
}

// TearDownSuite is run after all tests complete.
func (ts *TestSuite) TearDownSuite() {
	ts.db.Close()
}

func TestDatabaseSuite(t *testing.T) {
	flag.Parse()
	if !*runTests {
		t.Skip("Skipping database integration tests.")
	}

	suite.Run(t, new(TestSuite))
}
