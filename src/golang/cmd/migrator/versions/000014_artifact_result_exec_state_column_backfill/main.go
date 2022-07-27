package _00014_artifact_result_exec_state_column_backfill

import (
	"context"
	"github.com/aqueducthq/aqueduct/lib/database"
)

func Up(ctx context.Context, db database.Database) error {
	execStateInfos, err := getExecStateForEachArtifactResult(ctx, db)
	if err != nil {
		return err
	}

	for _, execStateInfo := range execStateInfos {
		if !execStateInfo.ExecState.IsNull {
			err = updateExecStateInArtifactResult(
				ctx,
				execStateInfo.ArtifactResultID,
				&execStateInfo.ExecState.ExecutionState,
				db,
			)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func Down(ctx context.Context, db database.Database) error {
	// This updates the status field with the exec state fetched from the operator result.
	// In the future, we can simply drop the status field without having to worry about downgrading. TODO(ENG-1453).
	execStateInfos, err := getExecStateForEachArtifactResult(ctx, db)
	if err != nil {
		return err
	}

	for _, execStateInfo := range execStateInfos {
		err = updateStatusInArtifactResult(
			ctx,
			execStateInfo.ArtifactResultID,
			execStateInfo.ExecState.Status,
			db,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
