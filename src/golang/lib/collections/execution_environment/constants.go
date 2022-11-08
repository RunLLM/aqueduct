package execution_environment

import (
	"fmt"
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

// Returns a joined string of all WorkflowDagResult columns prefixed by table name.
func allColumnsWithPrefix() string {
	return strings.Join(
		[]string{
			fmt.Sprintf("%s.%s", tableName, IdColumn),
			fmt.Sprintf("%s.%s", tableName, SpecColumn),
			fmt.Sprintf("%s.%s", tableName, HashColumn),
		},
		",",
	)
}
