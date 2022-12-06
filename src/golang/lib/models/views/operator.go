package views

import (
	"time"

	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	"github.com/google/uuid"
)

// Specifically used by GetDistinctLoadOperatorsByWorkflow
type LoadOperator struct {
	OperatorName    string         `db:"operator_name" json:"operator_name"`
	ModifiedAt      time.Time      `db:"modified_at" json:"modified_at"`
	IntegrationName string         `db:"integration_name" json:"integration_name"`
	IntegrationID   uuid.UUID      `db:"integration_id" json:"integration_id"`
	Service         integration.Service `db:"service" json:"service"`
	TableName       string         `db:"table_name" json:"object_name"`
	UpdateMode      string         `db:"update_mode" json:"update_mode"`
}
