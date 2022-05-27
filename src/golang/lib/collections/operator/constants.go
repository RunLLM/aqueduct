package operator

import (
	"strings"
)

const (
	tableName = "operator"

	// Operator table column names
	IdColumn          = "id"
	NameColumn        = "name"
	DescriptionColumn = "description"
	SpecColumn        = "spec"
)

// Returns a joined string of all Operator columns.
func allColumns() string {
	return strings.Join(
		[]string{
			IdColumn,
			NameColumn,
			DescriptionColumn,
			SpecColumn,
		},
		",",
	)
}
