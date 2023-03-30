package views

import (
	"fmt"
	"strings"

	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared/operator"
	"github.com/aqueducthq/aqueduct/lib/models/utils"
	"github.com/google/uuid"
)

const (
	OperatorNodeView    = "operator_node"
	OperatorNodeInputs  = "inputs"
	OperatorNodeOutputs = "outputs"
)

type OperatorNode struct {
	ID                     uuid.UUID      `db:"id" json:"id"`
	DagID                  uuid.UUID      `db:"dag_id" json:"dag_id"`
	Name                   string         `db:"name" json:"name"`
	Description            string         `db:"description" json:"description"`
	Spec                   operator.Spec  `db:"spec" json:"spec"`
	ExecutionEnvironmentID utils.NullUUID `db:"execution_environment_id" json:"execution_environment_id"`

	Inputs  []uuid.UUID `db:"inputs" json:"inputs"`
	Outputs []uuid.UUID `db:"outputs" json:"outputs"`
}

// OperatorNodeCols returns a comma-separated string of all Operator columns.
func OperatorNodeCols() string {
	return strings.Join(allOperatorNodeCols(), ",")
}

// OperatorNodeColsWithPrefix returns a comma-separated string of all
// operator columns prefixed by the table name.
func OperatorNodeColsWithPrefix() string {
	cols := allOperatorNodeCols()
	for i, col := range cols {
		cols[i] = fmt.Sprintf("%s.%s", OperatorNodeView, col)
	}

	return strings.Join(cols, ",")
}

func allOperatorNodeCols() []string {
	opNodeCols := models.AllOperatorCols()
	opNodeCols = append(
		opNodeCols,
		OperatorNodeInputs,
		OperatorNodeOutputs,
	)

	return opNodeCols
}
