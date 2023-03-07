package models

import (
	"fmt"
	"strings"

	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
)

const (
	ExecutionEnvironmentTable = "execution_environment"

	// ExecutionEnvironment table column names
	ExecutionEnvironmentID               = "id"
	ExecutionEnvironmentSpec             = "spec"
	ExecutionEnvironmentHash             = "hash"
	ExecutionEnvironmentGarbageCollected = "garbage_collected"
)

// A ExecutionEnvironment maps to the execution_environment table.
type ExecutionEnvironment struct {
	ID                         uuid.UUID                       `db:"id" json:"id"`
	Spec                       shared.ExecutionEnvironmentSpec `db:"spec" json:"spec"`
	Hash                       uuid.UUID                       `db:"hash" json:"hash"`
	GarbageCollectedDeprecated bool                            `db:"garbage_collected" json:"garbage_collected"`
}

// ExecutionEnvironmentCols returns a comma-separated string of all ExecutionEnvironment columns.
func ExecutionEnvironmentCols() string {
	return strings.Join(allExecutionEnvironmentCols(), ",")
}

// ExecutionEnvironmentColsWithPrefix returns a comma-separated string of all
// ExecutionEnvironment columns prefixed by the table name.
func ExecutionEnvironmentColsWithPrefix() string {
	cols := allExecutionEnvironmentCols()
	for i, col := range cols {
		cols[i] = fmt.Sprintf("%s.%s", ExecutionEnvironmentTable, col)
	}

	return strings.Join(cols, ",")
}

func allExecutionEnvironmentCols() []string {
	return []string{
		ExecutionEnvironmentID,
		ExecutionEnvironmentSpec,
		ExecutionEnvironmentHash,
		ExecutionEnvironmentGarbageCollected,
	}
}
