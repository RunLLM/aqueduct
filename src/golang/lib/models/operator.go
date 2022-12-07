package models

import (
	"strings"

	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
)

const (
	OperatorTable = "operator"

	// Operator table column names
	OperatorID          = "id"
	OperatorName        = "name"
	OperatorDescription = "description"
	OperatorSpec        = "spec"
)

// A Operator maps to the operator table.
type Operator struct {
	ID          uuid.UUID   `db:"id" json:"id"`
	Name        string      `db:"name" json:"name"`
	Description string      `db:"description" json:"description"`
	Spec        shared.Spec `db:"spec" json:"spec"`

	/* Fields not stored in DB */
	Inputs  []uuid.UUID `json:"inputs"`
	Outputs []uuid.UUID `json:"outputs"`
}

// OperatorCols returns a comma-separated string of all Operator columns.
func OperatorCols() string {
	return strings.Join(allOperatorCols(), ",")
}

func allOperatorCols() []string {
	return []string{
		OperatorID,
		OperatorName,
		OperatorDescription,
		OperatorSpec,
	}
}
