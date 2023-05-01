package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/aqueducthq/aqueduct/lib/models/shared"
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

// A DAGResult maps to the workflow_dag_result table.
type DAGResult struct {
	ID     uuid.UUID              `db:"id" json:"id"`
	DagID  uuid.UUID              `db:"workflow_dag_id" json:"workflow_dag_id"`
	Status shared.ExecutionStatus `db:"status" json:"status"`
	// TODO ENG-1701: deprecate `CreatedAt` field.
	CreatedAt time.Time                 `db:"created_at" json:"created_at"`
	ExecState shared.NullExecutionState `db:"execution_state" json:"execution_state"`
}

// DAGResultCols returns a comma-separated string of all DAGResult columns.
func DAGResultCols() string {
	return strings.Join(AllDAGResultCols(), ",")
}

// DAGResultColsWithPrefix returns a comma-separated string of all
// DAGResult columns prefixed by the table name.
func DAGResultColsWithPrefix() string {
	cols := AllDAGResultCols()
	for i, col := range cols {
		cols[i] = fmt.Sprintf("%s.%s", DAGResultTable, col)
	}

	return strings.Join(cols, ",")
}

func AllDAGResultCols() []string {
	return []string{
		DAGResultID,
		DAGResultDagID,
		DAGResultStatus,
		DAGResultCreatedAt,
		DAGResultExecState,
	}
}
