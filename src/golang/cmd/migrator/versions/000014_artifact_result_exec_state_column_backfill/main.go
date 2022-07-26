package _00014_artifact_result_exec_state_column_backfill

import (
	"context"
	"fmt"
	"github.com/aqueducthq/aqueduct/lib/database"
)

func Up(ctx context.Context, db database.Database) error {
	execStateInfos, err := getExecutionStateForEachArtifactID(ctx, db)
	if err != nil {
		return err
	}

	fmt.Println("HELLO: ", execStateInfos)

	return nil
}

func Down(ctx context.Context, db database.Database) error {
	// TODO: figure this out
	return nil
}
