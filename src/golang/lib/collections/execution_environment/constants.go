package execution_environment

import (
	"fmt"
	"strings"
)

const (
	tableName = "execution_environment"

	// ExecutionEnvironment table column names
	IdColumn               = "id"
	SpecColumn             = "spec"
	HashColumn             = "hash"
	GarbageCollectedColumn = "garbage_collected"
)

// Returns a joined string of all ExecutionEnvironment columns.
func allColumns() string {
	return strings.Join(
		[]string{
			IdColumn,
			SpecColumn,
			HashColumn,
			GarbageCollectedColumn,
		},
		",",
	)
}

// Returns a joined string of all ExecutionEnvironment columns
// prefixed by table name
func allColumnsWithPrefix() string {
	return strings.Join(
		[]string{
			fmt.Sprintf("%s.%s", tableName, IdColumn),
			fmt.Sprintf("%s.%s", tableName, SpecColumn),
			fmt.Sprintf("%s.%s", tableName, HashColumn),
			fmt.Sprintf("%s.%s", tableName, GarbageCollectedColumn),
		},
		",",
	)
}
