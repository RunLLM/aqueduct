package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
)

const (
	DagTable = "workflow_dag"

	// DAG column names
	DagID            = "id"
	DagWorkflowID    = "workflow_id"
	DagCreatedAt     = "created_at"
	DagStorageConfig = "storage_config"
	DagEngineConfig  = "engine_config"
)

// A DAG maps to the workflow_dag table.
type DAG struct {
	ID            uuid.UUID            `db:"id" json:"id"`
	WorkflowID    uuid.UUID            `db:"workflow_id" json:"workflow_id"`
	CreatedAt     time.Time            `db:"created_at" json:"created_at"`
	StorageConfig shared.StorageConfig `db:"storage_config" json:"storage_config"`

	// Sets the default engine for DAG execution. Can be overridden by the operator spec.
	EngineConfig shared.EngineConfig `db:"engine_config" json:"engine_config"`

	/* Field not stored in DB */
	Metadata  *Workflow              `json:"metadata"`
	Operators map[uuid.UUID]Operator `json:"operators,omitempty"`
	Artifacts map[uuid.UUID]Artifact `json:"artifacts,omitempty"`
}

// DAGCols returns a comma-separated string of all DAG columns.
func DAGCols() string {
	return strings.Join(AllDAGCols(), ",")
}

// DAGColsWithPrefix returns a comma-separated string of all
// DAG columns prefixed by the table name.
func DAGColsWithPrefix() string {
	cols := AllDAGCols()
	for i, col := range cols {
		cols[i] = fmt.Sprintf("%s.%s", DagTable, col)
	}

	return strings.Join(cols, ",")
}

func AllDAGCols() []string {
	return []string{
		DagID,
		DagWorkflowID,
		DagCreatedAt,
		DagStorageConfig,
		DagEngineConfig,
	}
}
