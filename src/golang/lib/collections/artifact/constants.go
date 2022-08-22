package artifact

import (
	"strings"
)

const (
	tableName = "artifact"

	// Artifact table column names
	IdColumn          = "id"
	NameColumn        = "name"
	DescriptionColumn = "description"
	TypeColumn        = "type"
)

// Returns a joined string of all Artifact columns.
func allColumns() string {
	return strings.Join(
		[]string{
			IdColumn,
			NameColumn,
			DescriptionColumn,
			TypeColumn,
		},
		",",
	)
}
