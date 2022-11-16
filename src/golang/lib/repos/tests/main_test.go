package tests

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/suite"
)

var runTests = flag.Bool("database", false, "If this flag is set, the database integration tests will be run.")

// TestDatabaseSuite is the entrypoint for all database integration tests
// in this package.
func TestDatabaseSuite(t *testing.T) {
	flag.Parse()
	if !*runTests {
		t.Skip("Skipping database integration tests.")
	}

	suite.Run(t, new(TestSuite))
}
