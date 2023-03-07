package models

import (
	"strings"
)

const (
	// SchemaVersion is the current database schema version.
	// This is the source of truth for the required schema version
	// for both the server and executor. This value MUST be updated
	// when a new schema change is added.
	CurrentSchemaVersion = 24

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
