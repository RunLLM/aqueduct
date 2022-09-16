package tests

import (
	"context"
	"testing"
	"time"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_result"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// seedWorkflowDagResultWithDags populates the workflow_dag_result table with count
// workflow dag results where workflow_dag_id is set to the values provided in workflowDagIds.
func seedWorkflowDagResultWithDags(
	t *testing.T,
	count int,
	workflowDagIds []uuid.UUID,
) []workflow_dag_result.WorkflowDagResult {
	require.Equal(t, count, len(workflowDagIds))

	now := time.Now()
	dagResults := make([]workflow_dag_result.WorkflowDagResult, 0, count)
	for i := 0; i < count; i++ {
		testDagResult, err := writers.workflowDagResultWriter.CreateWorkflowDagResult(
			context.Background(),
			workflowDagIds[i],
			&shared.ExecutionState{
				Status: shared.PendingExecutionStatus,
				Timestamps: &shared.ExecutionTimestamps{
					PendingAt: &now,
				},
			},
			db,
		)
		require.Nil(t, err)

		dagResults = append(dagResults, *testDagResult)
	}

	require.Len(t, dagResults, count)

	return dagResults
}

func TestCreateWorkflowDagResult(t *testing.T) {
	defer resetDatabase(t)

	dags := seedWorkflowDag(t, 1)
	now := time.Now()
	execState := &shared.ExecutionState{
		Status: shared.PendingExecutionStatus,
		Timestamps: &shared.ExecutionTimestamps{
			PendingAt: &now,
		},
	}

	expectedDagResult := &workflow_dag_result.WorkflowDagResult{
		WorkflowDagId: dags[0].Id,
		Status:        shared.PendingExecutionStatus,
		ExecState: shared.NullExecutionState{
			ExecutionState: *execState,
		},
	}

	actualDagResult, err := writers.workflowDagResultWriter.CreateWorkflowDagResult(
		context.Background(),
		expectedDagResult.WorkflowDagId,
		execState,
		db,
	)
	require.Nil(t, err)
	require.NotEqual(t, uuid.Nil, actualDagResult.Id)

	expectedDagResult.Id = actualDagResult.Id
	expectedDagResult.CreatedAt = actualDagResult.CreatedAt

	requireDeepEqual(t, expectedDagResult, actualDagResult)
}

func TestGetKOffsetWorkflowDagResultsByWorkflowId(t *testing.T) {
	defer resetDatabase(t)

	numWorkflowDags := 1
	numWorkflowDagResults := 5

	testWorkflowDags := seedWorkflowDag(t, numWorkflowDags)
	testDagIds := randWorkflowDagIdsFromList(numWorkflowDagResults, testWorkflowDags)

	seedWorkflowDagResultWithDags(t, numWorkflowDagResults, time.Now(), testDagIds)

	kOffsetWorkflowDagResults, err := readers.workflowDagResultReader.GetKOffsetWorkflowDagResultsByWorkflowId(
		context.Background(),
		testWorkflowDags[0].WorkflowId,
		1,
		db,
	)

	require.Nil(t, err)
	require.Equal(t, 4, len(kOffsetWorkflowDagResults))
}
