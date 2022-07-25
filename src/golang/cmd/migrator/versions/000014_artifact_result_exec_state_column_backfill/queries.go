package _00014_artifact_result_exec_state_column_backfill

import (
	"context"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/google/uuid"
)

type artifactOperatorExecState struct {
	OpResultID       uuid.UUID                 `db:"operator_result_id"`
	ArtifactResultID uuid.UUID                 `db:"artifact_result_id"`
	ExecState        shared.NullExecutionState `db:"execution_state"`
}

func getExecutionStateForEachArtifactID(ctx context.Context, db database.Database) ([]artifactOperatorExecState, error) {
	query := `
		SELECT operator_result.id as operator_result_id, 
		artifact_result.id as artifact_result_id, 
		operator_result.execution_state
		FROM operator_result
		INNER JOIN operator ON operator_result.operator_id=operator.id
		INNER JOIN workflow_dag_edge ON operator.id=workflow_dag_edge.from_id
		INNER JOIN artifact ON workflow_dag_edge.to_id=artifact.id
		INNER JOIN artifact_result ON artifact.id=artifact_result.artifact_id;
	`

	var info []artifactOperatorExecState
	err := db.Query(ctx, &info, query)
	return info, err
}
