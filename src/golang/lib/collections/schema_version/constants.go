package schema_version

import "strings"

const (
	tableName = "schema_version"

	// SchemaVersion table column names
	VersionColumn = "version"
	DirtyColumn   = "dirty"
	NameColumn    = "name"
)

// Returns a joined string of all SchemaVersion columns.
func allColumns() string {
	return strings.Join(
		[]string{
			VersionColumn,
			DirtyColumn,
			NameColumn,
		},
		",",
	)
}
