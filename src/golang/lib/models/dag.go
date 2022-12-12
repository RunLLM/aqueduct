package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/google/uuid"
)

const (
	DagTable = "workflow_dag"

	// DAG column names
	DagID                     = "id"
	DagWorkflowID             = "workflow_id"
	DagCreatedAt              = "created_at"
	DagStorageConfig          = "storage_config"
	DeprecatedDagEngineConfig = "engine_config"
)

// A DAG maps to the workflow_dag table.
type DAG struct {
	ID            uuid.UUID            `db:"id" json:"id"`
	WorkflowID    uuid.UUID            `db:"workflow_id" json:"workflow_id"`
	CreatedAt     time.Time            `db:"created_at" json:"created_at"`
	StorageConfig shared.StorageConfig `db:"storage_config" json:"storage_config"`
	EngineConfig  shared.EngineConfig  `db:"engine_config" json:"engine_config"`

	/* Field not stored in DB */
	Metadata  *Workflow              `json:"metadata"`
	Operators map[uuid.UUID]Operator `json:"operators,omitempty"`
	Artifacts map[uuid.UUID]Artifact `json:"artifacts,omitempty"`
}

// DAGCols returns a comma-separated string of all DAG columns.
func DAGCols() string {
	return strings.Join(allDAGCols(), ",")
}

// DAGColsWithPrefix returns a comma-separated string of all
// DAG columns prefixed by the table name.
func DAGColsWithPrefix() string {
	cols := allDAGCols()
	for i, col := range cols {
		cols[i] = fmt.Sprintf("%s.%s", DagTable, col)
	}

	return strings.Join(cols, ",")
}

func allDAGCols() []string {
	return []string{
		DagID,
		DagWorkflowID,
		DagCreatedAt,
		DagStorageConfig,
		DeprecatedDagEngineConfig,
	}
}
