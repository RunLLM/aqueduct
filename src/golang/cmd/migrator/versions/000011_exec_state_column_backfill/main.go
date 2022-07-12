package _000011_exec_state_column_backfill

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
)

func Up(ctx context.Context, db database.Database) error {
	opResults, err := getOpResultsWithMetadata(ctx, db)
	if err != nil {
		return err
	}

	for _, opResult := range opResults {
		if !opResult.Metadata.IsNull {
			userLogs := Logs{}
			stdout, ok := opResult.Metadata.Logs["stdout"]
			if ok {
				(&userLogs).Stdout = stdout
			}

			stderr, ok := opResult.Metadata.Logs["stderr"]
			if ok {
				(&userLogs).StdErr = stderr
			}

			execState := ExecutionState{
				Status: opResult.Status,
				Error: &Error{
					Context: opResult.Metadata.Error,
				},
				UserLogs: &userLogs,
			}

			err = updateExecState(ctx, opResult.Id, &execState, db)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func Down(ctx context.Context, db database.Database) error {
	opResults, err := getOpResultsWithExecState(ctx, db)
	if err != nil {
		return err
	}

	for _, opResult := range opResults {
		if !opResult.ExecState.IsNull {
			Logs := map[string]string{}
			if opResult.ExecState.UserLogs != nil {
				Logs["stdout"] = opResult.ExecState.UserLogs.Stdout
				Logs["stderr"] = opResult.ExecState.UserLogs.StdErr
			}

			errStr := ""
			if opResult.ExecState.Error != nil {
				errStr = opResult.ExecState.Error.Context
			}

			metadata := Metadata{Logs: Logs, Error: errStr}

			err = updateMetadata(ctx, opResult.Id, &metadata, db)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
