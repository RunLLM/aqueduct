package tests

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/suite"
)

var runTests = flag.Bool("database", false, "If this flag is set, the database resource tests will be run.")

// TestDatabaseSuite is the entrypoint for all database resource tests
// in this package.
func TestDatabaseSuite(t *testing.T) {
	flag.Parse()
	if !*runTests {
		t.Skip("Skipping database resource tests.")
	}

	suite.Run(t, new(TestSuite))
}
