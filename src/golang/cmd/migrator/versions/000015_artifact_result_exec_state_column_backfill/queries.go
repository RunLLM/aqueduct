package _000015_artifact_result_exec_state_column_backfill

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/google/uuid"
)

type artifactOperatorExecState struct {
	ArtifactResultID uuid.UUID                 `db:"id"`
	ExecState        shared.NullExecutionState `db:"execution_state"`
}

func getExecStateForEachArtifactResult(ctx context.Context, db database.Database) ([]artifactOperatorExecState, error) {
	query := `
		SELECT artifact_result.id, operator_result.execution_state
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

func updateExecStateInArtifactResult(
	ctx context.Context,
	id uuid.UUID,
	execState *shared.ExecutionState,
	db database.Database,
) error {
	changes := map[string]interface{}{
		"execution_state": execState,
	}
	return repos.UpdateRecord(ctx, changes, "artifact_result", "id", id, db)
}

func updateStatusInArtifactResult(
	ctx context.Context,
	id uuid.UUID,
	status shared.ExecutionStatus,
	db database.Database,
) error {
	changes := map[string]interface{}{
		"status": status,
	}
	return repos.UpdateRecord(ctx, changes, "artifact_result", "id", id, db)
}
