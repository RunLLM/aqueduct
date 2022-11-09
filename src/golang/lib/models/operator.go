package models

import (
	"strings"
	"time"

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

// Specifically used by GetDistinctLoadOperatorsByWorkflow
type LoadOperator struct {
	OperatorName    string         `db:"operator_name" json:"operator_name"`
	ModifiedAt      time.Time      `db:"modified_at" json:"modified_at"`
	IntegrationName string         `db:"integration_name" json:"integration_name"`
	IntegrationID   uuid.UUID      `db:"integration_id" json:"integration_id"`
	Service         shared.Service `db:"service" json:"service"`
	TableName       string         `db:"table_name" json:"object_name"`
	UpdateMode      string         `db:"update_mode" json:"update_mode"`
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