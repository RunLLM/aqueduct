package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/google/uuid"
)

const (
	DAGResultTable = "workflow_dag_result"

	// DAGResult column names
	DAGResultID        = "id"
	DAGResultDagID     = "workflow_dag_id"
	DAGResultStatus    = "status"
	DAGResultCreatedAt = "created_at"
	DAGResultExecState = "execution_state"
)

type DAGResult struct {
	ID    uuid.UUID `db:"id" json:"id"`
	DagID uuid.UUID `db:"workflow_dag_id" json:"workflow_dag_id"`
	// TODO: Refactor once Operator refactor is merged
	Status shared.ExecutionStatus `db:"status" json:"status"`
	// TODO ENG-1701: deprecate `CreatedAt` field.
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	// TODO: Refactor once Operator refactor is merged
	ExecState shared.NullExecutionState `db:"execution_state" json:"execution_state"`
}

// DAGResultCols returns a comma-separated string of all DAGResult columns.
func DAGResultCols() string {
	return strings.Join(allDAGResultCols(), ",")
}

// DAGResultColsWithPrefix returns a comma-separated string of all
// DAGResult columns prefixed by the table name.
func DAGResultColsWithPrefix() string {
	cols := allDAGResultCols()
	for i, col := range cols {
		cols[i] = fmt.Sprintf("%s.%s", DAGResultTable, col)
	}

	return strings.Join(cols, ",")
}

func allDAGResultCols() []string {
	return []string{
		DAGResultID,
		DAGResultDagID,
		DAGResultStatus,
		DAGResultCreatedAt,
		DAGResultExecState,
	}
}
