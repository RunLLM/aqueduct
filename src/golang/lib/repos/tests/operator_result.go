package tests

import (
	"time"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/views"
	"github.com/aqueducthq/aqueduct/lib/models/utils"
)

func (ts *TestSuite) TestOperatorResult_Get() {
	operatorResults := ts.seedOperatorResultForDAGAndOperator(1, uuid.New(), uuid.New())
	expectedOperatorResult := operatorResults[0]

	actualOperatorResult, err := ts.operatorResult.Get(ts.ctx, expectedOperatorResult.ID, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqual(ts.T(), expectedOperatorResult, *actualOperatorResult)
}

func (ts *TestSuite) TestOperatorResult_GetBatch() {
	expectedOperatorResults := ts.seedOperatorResultForDAGAndOperator(3, uuid.New(), uuid.New())

	IDs := make([]uuid.UUID, 0, len(expectedOperatorResults))
	for _, expectedOperatorResult := range expectedOperatorResults {
		IDs = append(IDs, expectedOperatorResult.ID)
	}

	actualOperatorResults, err := ts.operatorResult.GetBatch(ts.ctx, IDs, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqualOperatorResults(ts.T(), expectedOperatorResults, actualOperatorResults)
}

func (ts *TestSuite) TestOperatorResult_GetByDAGResultAndOperator() {
	operatorResults := ts.seedOperatorResultForDAGAndOperator(1, uuid.New(), uuid.New())
	expectedOperatorResult := operatorResults[0]

	actualOperatorResult, err := ts.operatorResult.GetByDAGResultAndOperator(ts.ctx, expectedOperatorResult.DAGResultID, expectedOperatorResult.OperatorID, ts.DB)
	require.Nil(ts.T(), err)
	// ExecState's timestamps is set equal since the timestamps will not match due to the fact
	// that they are pointers.
	expectedOperatorResult.ExecState.ExecutionState.Timestamps.PendingAt = actualOperatorResult.ExecState.ExecutionState.Timestamps.PendingAt
	requireDeepEqual(ts.T(), &expectedOperatorResult, actualOperatorResult)
}

func (ts *TestSuite) TestOperatorResult_GetByDAGResultBatch() {
	operatorId := uuid.New()
	expectedOperatorResultsA := ts.seedOperatorResultForDAGAndOperator(3, uuid.New(), operatorId)
	expectedOperatorResultsB := ts.seedOperatorResultForDAGAndOperator(3, uuid.New(), operatorId)
	_ = ts.seedOperatorResultForDAGAndOperator(3, uuid.New(), operatorId)

	actualOperatorResults, err := ts.operatorResult.GetByDAGResultBatch(ts.ctx, []uuid.UUID{expectedOperatorResultsA[0].DAGResultID, expectedOperatorResultsB[0].DAGResultID}, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqualOperatorResults(ts.T(), append(expectedOperatorResultsA, expectedOperatorResultsB...), actualOperatorResults)
}

func (ts *TestSuite) TestOperatorResult_GetCheckStatusByArtifactBatch() {
	expectedOperatorResults, operator, artifactID := ts.seedOperatorResult(3, operator.CheckType)

	expectedOperatorResultStatuses := make([]views.OperatorResultStatus, 0, len(expectedOperatorResults))
	for _, expectedOperatorResult := range expectedOperatorResults {
		expectedOperatorResultStatus := views.OperatorResultStatus{
			ArtifactID: artifactID,
			Metadata: &expectedOperatorResult.ExecState.ExecutionState,
			DAGResultID: expectedOperatorResult.DAGResultID,
			OperatorName: utils.NullString{
				String: operator.Name,
				IsNull: false,
			},
		}
		expectedOperatorResultStatuses = append(expectedOperatorResultStatuses, expectedOperatorResultStatus)
	}
	
	actualOperatorResultStatuses, err := ts.operatorResult.GetCheckStatusByArtifactBatch(ts.ctx, []uuid.UUID{artifactID}, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqualOperatorResultStatuses(ts.T(), expectedOperatorResultStatuses, actualOperatorResultStatuses)
}

func (ts *TestSuite) TestOperatorResult_GetStatusByDAGResultAndArtifactBatch() {
	artifactIDA := uuid.New()
	artifactIDB := uuid.New()

	dagResultIDA := uuid.New()
	dagResultIDB := uuid.New()
	dagResultIDC := dagResultIDB
	dagResultIDD := dagResultIDB

	operatorA := ts.seedOperatorAndDAGOperatorToArtifact(artifactIDA, dagResultIDA, operator.FunctionType)
	operatorB := ts.seedOperatorAndDAGOperatorToArtifact(artifactIDB, dagResultIDB, operator.FunctionType)
	operatorC := ts.seedOperatorAndDAGOperatorToArtifact(artifactIDA, dagResultIDC, operator.FunctionType)
	operatorD := ts.seedOperatorAndDAGOperatorToArtifact(uuid.New(), dagResultIDD, operator.FunctionType)
	
	expectedOperatorResultsA := ts.seedOperatorResultForDAGAndOperator(3, dagResultIDA, operatorA.ID)
	expectedOperatorResultsB := ts.seedOperatorResultForDAGAndOperator(3, dagResultIDB, operatorB.ID)
	expectedOperatorResultsC := ts.seedOperatorResultForDAGAndOperator(3, dagResultIDC, operatorC.ID)
	_ = ts.seedOperatorResultForDAGAndOperator(3, dagResultIDD, operatorD.ID)

	expectedOperatorResultStatuses := make([]views.OperatorResultStatus, 0, len(expectedOperatorResultsA)+len(expectedOperatorResultsB)+len(expectedOperatorResultsC))
	for _, expectedOperatorResult := range expectedOperatorResultsA {
		expectedOperatorResultStatus := views.OperatorResultStatus{
			ArtifactID: artifactIDA,
			Metadata: &expectedOperatorResult.ExecState.ExecutionState,
			DAGResultID: expectedOperatorResult.DAGResultID,
			OperatorName: utils.NullString{
				String: "",
				IsNull: true,
			},
		}
		expectedOperatorResultStatuses = append(expectedOperatorResultStatuses, expectedOperatorResultStatus)
	}

	for _, expectedOperatorResult := range expectedOperatorResultsB {
		expectedOperatorResultStatus := views.OperatorResultStatus{
			ArtifactID: artifactIDB,
			Metadata: &expectedOperatorResult.ExecState.ExecutionState,
			DAGResultID: expectedOperatorResult.DAGResultID,
			OperatorName: utils.NullString{
				String: "",
				IsNull: true,
			},
		}
		expectedOperatorResultStatuses = append(expectedOperatorResultStatuses, expectedOperatorResultStatus)
	}

	for _, expectedOperatorResult := range expectedOperatorResultsC {
		expectedOperatorResultStatus := views.OperatorResultStatus{
			ArtifactID: artifactIDB,
			Metadata: &expectedOperatorResult.ExecState.ExecutionState,
			DAGResultID: expectedOperatorResult.DAGResultID,
			OperatorName: utils.NullString{
				String: "",
				IsNull: true,
			},
		}
		expectedOperatorResultStatuses = append(expectedOperatorResultStatuses, expectedOperatorResultStatus)
	}
	
	actualOperatorResultStatuses, err := ts.operatorResult.GetStatusByDAGResultAndArtifactBatch(ts.ctx, []uuid.UUID{dagResultIDA, dagResultIDB}, []uuid.UUID{artifactIDA, artifactIDB}, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqualOperatorResultStatuses(ts.T(), expectedOperatorResultStatuses, actualOperatorResultStatuses)
}

func (ts *TestSuite) TestOperatorResult_Create() {
	dagResultID := uuid.New()
	operatorID := uuid.New()
	now := time.Now()
	execState := shared.ExecutionState{
		Status: shared.PendingExecutionStatus,
		Timestamps: &shared.ExecutionTimestamps{
			PendingAt: &now,
		},
	}
	expectedOperatorResult := &models.OperatorResult{
		DAGResultID: dagResultID,
		OperatorID: operatorID,
		Status: execState.Status,
		ExecState: shared.NullExecutionState{
			ExecutionState: execState,
			IsNull: false,
		},
	}
	actualOperatorResult, err := ts.operatorResult.Create(
		ts.ctx,
		dagResultID,
		operatorID,
		&execState,
		ts.DB,
	)
	require.Nil(ts.T(), err)
	expectedOperatorResult.ID = actualOperatorResult.ID
	// ExecState's timestamps is set equal since the timestamps will not match due to the fact
	// that they are pointers.
	expectedOperatorResult.ExecState.ExecutionState.Timestamps.PendingAt = actualOperatorResult.ExecState.ExecutionState.Timestamps.PendingAt
	requireDeepEqual(ts.T(), expectedOperatorResult, actualOperatorResult)
}

func (ts *TestSuite) TestOperatorResult_Delete() {
	operatorResults := ts.seedOperatorResultForDAGAndOperator(1, uuid.New(), uuid.New())

	err := ts.operatorResult.Delete(ts.ctx, operatorResults[0].ID, ts.DB)
	require.Nil(ts.T(), err)
}

func (ts *TestSuite) TestOperatorResult_DeleteBatch() {
	operatorResults := ts.seedOperatorResultForDAGAndOperator(3, uuid.New(), uuid.New())

	IDs := make([]uuid.UUID, 0, len(operatorResults))
	for _, operatorResult := range operatorResults {
		IDs = append(IDs, operatorResult.ID)
	}

	err := ts.operatorResult.DeleteBatch(ts.ctx, IDs, ts.DB)
	require.Nil(ts.T(), err)
}

func (ts *TestSuite) TestOperatorResult_Update() {
	operatorResults := ts.seedOperatorResultForDAGAndOperator(1, uuid.New(), uuid.New())
	expectedOperatorResult := operatorResults[0]

	execState := shared.NullExecutionState{
		ExecutionState: shared.ExecutionState{
			UserLogs: &shared.Logs{
				Stdout: randString(10),
				StdErr: randString(10),
			},
			Status: shared.UnknownExecutionStatus,
		},
		IsNull: false,
	}

	changes := map[string]interface{}{
		models.OperatorResultExecState:   &execState,
	}

	actualOperatorResult, err := ts.operatorResult.Update(ts.ctx, expectedOperatorResult.ID, changes, ts.DB)
	require.Nil(ts.T(), err)

	expectedOperatorResult.ExecState = execState

	requireDeepEqual(ts.T(), &expectedOperatorResult, actualOperatorResult)
}
