package tests

import (
	"time"

	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func (ts *TestSuite) TestDAGResult_Get() {
	dagResults := ts.seedDAGResult(1)
	expexctedDAGResult := dagResults[0]

	actualDAGResult, err := ts.dagResult.Get(ts.ctx, expexctedDAGResult.ID, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqual(ts.T(), expexctedDAGResult, *actualDAGResult)
}

func (ts *TestSuite) TestDAGResult_GetBatch() {
	expectedDAGResults := ts.seedDAGResult(3)

	IDs := make([]uuid.UUID, 0, len(expectedDAGResults))
	for _, expectedDAGResult := range expectedDAGResults {
		IDs = append(IDs, expectedDAGResult.ID)
	}

	actualDAGResults, err := ts.dagResult.GetBatch(ts.ctx, IDs, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqualDAGResults(ts.T(), expectedDAGResults, actualDAGResults)
}

func (ts *TestSuite) TestDAGResult_GetByWorkflow() {
	dags := ts.seedDAG(1)
	dag := dags[0]

	expectedDAGResults := ts.seedDAGResultWithDAG(2, []uuid.UUID{dag.ID, dag.ID})

	actualDAGResults, err := ts.dagResult.GetByWorkflow(ts.ctx, dag.WorkflowID, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqualDAGResults(ts.T(), expectedDAGResults, actualDAGResults)
}

func (ts *TestSuite) TestDAGResult_GetKOffsetByWorkflow() {
	dags := ts.seedDAG(1)
	dag := dags[0]

	expectedDAGResults := ts.seedDAGResultWithDAG(
		2,
		[]uuid.UUID{dag.ID, dag.ID},
	)

	// Seed 2 more DAGResults with a later CreatedAt
	ts.seedDAGResultWithDAG(
		2,
		[]uuid.UUID{dag.ID, dag.ID},
	)

	actualDAGResults, err := ts.dagResult.GetKOffsetByWorkflow(
		ts.ctx,
		dag.WorkflowID, 2,
		ts.DB,
	)
	require.Nil(ts.T(), err)
	requireDeepEqualDAGResults(ts.T(), expectedDAGResults, actualDAGResults)
}

func (ts *TestSuite) TestDAGResult_Create() {
	dags := ts.seedDAG(1)
	dag := dags[0]

	now := time.Now()
	expectedDAGResult := &models.DAGResult{
		DagID: dag.ID,
		ExecState: shared.NullExecutionState{
			IsNull: false,
			ExecutionState: shared.ExecutionState{
				Status: shared.PendingExecutionStatus,
				Timestamps: &shared.ExecutionTimestamps{
					PendingAt: &now,
				},
			},
		},
	}

	actualDAGResult, err := ts.dagResult.Create(
		ts.ctx,
		expectedDAGResult.DagID,
		&expectedDAGResult.ExecState.ExecutionState,
		ts.DB,
	)
	require.Nil(ts.T(), err)

	require.NotEqual(ts.T(), uuid.Nil, actualDAGResult.ID)

	expectedDAGResult.ID = actualDAGResult.ID
	expectedDAGResult.Status = actualDAGResult.Status
	// ExecState is set equal since the timestamps will not match due to the fact
	// that they are pointers.
	expectedDAGResult.ExecState = actualDAGResult.ExecState
	expectedDAGResult.CreatedAt = actualDAGResult.CreatedAt

	requireDeepEqual(ts.T(), expectedDAGResult, actualDAGResult)
}

func (ts *TestSuite) TestDAGResult_Delete() {
	dagResults := ts.seedDAGResult(1)
	dagResult := dagResults[0]

	err := ts.dagResult.Delete(ts.ctx, dagResult.ID, ts.DB)
	require.Nil(ts.T(), err)
}

func (ts *TestSuite) TestDAGResult_DeleteBatch() {
	dagResults := ts.seedDAGResult(2)
	IDs := []uuid.UUID{dagResults[0].ID, dagResults[1].ID}

	err := ts.dagResult.DeleteBatch(ts.ctx, IDs, ts.DB)
	require.Nil(ts.T(), err)
}

func (ts *TestSuite) TestDAGResult_Update() {
	dagResults := ts.seedDAGResult(1)
	dagResult := dagResults[0]

	newExecState := shared.ExecutionState{
		Status:     shared.SucceededExecutionStatus,
		Timestamps: dagResult.ExecState.Timestamps,
	}

	changes := map[string]interface{}{
		models.DAGResultExecState: &newExecState,
	}

	newDAGResult, err := ts.dagResult.Update(ts.ctx, dagResult.ID, changes, ts.DB)
	require.Nil(ts.T(), err)

	requireDeepEqual(ts.T(), newExecState, newDAGResult.ExecState.ExecutionState)
}
