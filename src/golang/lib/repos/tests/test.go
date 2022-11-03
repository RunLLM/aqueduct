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

type TestSuite struct {
	suite.Suite

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
