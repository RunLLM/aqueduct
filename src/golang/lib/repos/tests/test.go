package tests

import (
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/repos/sqlite"
	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	suite.Suite

	// List of all repos
	workflow repos.Workflow

	db database.Database
}

// SetupSuite is run only once before all tests. It initializes the database
// connection and creates the repos.
func (ts *TestSuite) SetupSuite() {
	db, err := database.NewSqliteInMemoryDatabase(&database.SqliteConfig{})
	if err != nil {
		ts.T().Errorf("Unable to create SQLite client: %v", err)
	}

	ts.db = db

	// Initialize repos
	ts.workflow = sqlite.NewWorklowRepo()
}
