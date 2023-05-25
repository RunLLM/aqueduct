package models

import (
	"strings"

	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
)

const (
	StorageMigrationTable = "storage_migration"

	// StorageMigration table column names
	StorageMigrationID = "id"

	// If null, the resource should be interpreted as the local filesystem.
	StorageMigrationDestResourceID = "dest_integration_id"
	StorageMigrationExecutionState = "execution_state"

	// This column must have at most one of these rows set to true.
	// Indicates what storage layer the server is currently using.
	// Equivalent to the result of the last successful migration.
	StorageMigrationCurrent = "current"
)

// A StorageMigration maps to the storage_migration table.
type StorageMigration struct {
	ID             uuid.UUID             `db:"id" json:"id"`
	DestResourceID uuid.UUID             `db:"dest_integration_id" json:"dest_integration_id"`
	ExecState      shared.ExecutionState `db:"execution_state" json:"execution_state"`
	Current        bool                  `db:"current" json:"current"`
}

func StorageMigrationCols() string {
	return strings.Join(allStorageMigrationCols(), ",")
}

func allStorageMigrationCols() []string {
	return []string{
		StorageMigrationID,
		StorageMigrationDestResourceID,
		StorageMigrationExecutionState,
		StorageMigrationCurrent,
	}
}
