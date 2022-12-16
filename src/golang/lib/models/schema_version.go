package models

import (
	"strings"
)

const (
	SchemaVersionTable = "schema_version"

	// SchemaVersion table column names
	SchemaVersionVersion = "version"
	SchemaVersionDirty   = "dirty"
	SchemaVersionName    = "name"
)

// A OperatorResult maps to the operator_result table.
type SchemaVersion struct {
	Version int64  `db:"version" json:"version"`
	Dirty   bool   `db:"dirty" json:"dirty"`
	Name    string `db:"name" json:"name`
}

// SchemaVersionCols returns a comma-separated string of all SchemaVersion columns.
func SchemaVersionCols() string {
	return strings.Join(allSchemaVersionCols(), ",")
}

func allSchemaVersionCols() []string {
	return []string{
		SchemaVersionVersion,
		SchemaVersionDirty,
		SchemaVersionName,
	}
}
