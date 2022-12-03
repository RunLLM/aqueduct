package models

import (
	"strings"

	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
)

const (
	DAGEdgeTable = "workflow_dag_edge"

	// DAGEdge column names
	DAGEdgeDagID  = "workflow_dag_id"
	DAGEdgeType   = "type"
	DAGEdgeFromID = "from_id"
	DAGEdgeToID   = "to_id"
	DAGEdgeIdx    = "idx"
)

// A DAGEdge maps to the workflow_dag_edge table.
type DAGEdge struct {
	DagID  uuid.UUID          `db:"workflow_dag_id" json:"workflow_dag_id"`
	Type   shared.DAGEdgeType `db:"type" json:"type"`
	FromID uuid.UUID          `db:"from_id" json:"from_id"`
	ToID   uuid.UUID          `db:"to_id" json:"to_id"`
	Idx    int16              `db:"idx" json:"idx"`
}

// DAGEdgeCols returns a comma-separated string of all DAGEdge columns.
func DAGEdgeCols() string {
	return strings.Join(allDAGEdgeCols(), ",")
}

func allDAGEdgeCols() []string {
	return []string{
		DAGEdgeDagID,
		DAGEdgeType,
		DAGEdgeFromID,
		DAGEdgeToID,
		DAGEdgeIdx,
	}
}
