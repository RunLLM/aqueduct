package execution_environment

import (
	"strings"
)

const (
	tableName = "execution_environment"

	// ExecutionEnvironment table column names
	IdColumn   = "id"
	SpecColumn = "spec"
	HashColumn = "hash"
)

// Returns a joined string of all ArtifactResult columns.
func allColumns() string {
	return strings.Join(
		[]string{
			IdColumn,
			SpecColumn,
			HashColumn,
		},
		",",
	)
}
