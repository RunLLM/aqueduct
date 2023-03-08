package models

import (
	"fmt"
	"strings"

	"github.com/aqueducthq/aqueduct/lib/models/shared/operator"
	"github.com/aqueducthq/aqueduct/lib/models/utils"
	"github.com/google/uuid"
)

const (
	OperatorTable = "operator"

	// Operator table column names
	OperatorID                     = "id"
	OperatorName                   = "name"
	OperatorDescription            = "description"
	OperatorSpec                   = "spec"
	OperatorExecutionEnvironmentID = "execution_environment_id"
)

// A Operator maps to the operator table.
type Operator struct {
	ID                     uuid.UUID      `db:"id" json:"id"`
	Name                   string         `db:"name" json:"name"`
	Description            string         `db:"description" json:"description"`
	Spec                   operator.Spec  `db:"spec" json:"spec"`
	ExecutionEnvironmentID utils.NullUUID `db:"execution_environment_id" json:"execution_environment_id"`

	/* Fields not stored in DB */
	Inputs  []uuid.UUID `json:"inputs"`
	Outputs []uuid.UUID `json:"outputs"`
}

// OperatorCols returns a comma-separated string of all Operator columns.
func OperatorCols() string {
	return strings.Join(allOperatorCols(), ",")
}

// OperatorColsWithPrefix returns a comma-separated string of all
// operator columns prefixed by the table name.
func OperatorColsWithPrefix() string {
	cols := allOperatorCols()
	for i, col := range cols {
		cols[i] = fmt.Sprintf("%s.%s", OperatorTable, col)
	}

	return strings.Join(cols, ",")
}

func allOperatorCols() []string {
	return []string{
		OperatorID,
		OperatorName,
		OperatorDescription,
		OperatorSpec,
		OperatorExecutionEnvironmentID,
	}
}
