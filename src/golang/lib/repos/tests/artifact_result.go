package tests

import (
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func (ts *TestSuite) TestArtifactResult_Get() {
	artifactResults, _, _, _ := ts.seedArtifactResult(1)
	expectedArtifactResult := &artifactResults[0]

	actualArtifactResult, err := ts.artifact_result.Get(ts.ctx, expectedArtifactResult.ID, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqual(ts.T(), expectedArtifactResult, actualArtifactResult)
}

func (ts *TestSuite) TestArtifactResult_GetBatch() {
	expectedArtifactResults, _, _, _  := ts.seedArtifactResult(3)

	IDs := make([]uuid.UUID, 0, len(expectedArtifactResults))
	for _, expectedArtifactResult := range expectedArtifactResults {
		IDs = append(IDs, expectedArtifactResult.ID)
	}

	actualArtifactResults, err := ts.artifact_result.GetBatch(ts.ctx, IDs, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqualArtifacts(ts.T(), expectedArtifactResults, actualArtifactResults)
}

func (ts *TestSuite) TestArtifactResult_GetByArtifact() {
	expectedArtifactResults, _, _, _  := ts.seedArtifactResult(3)

	actualArtifactResults, err := ts.artifact_result.GetByArtifact(ts.ctx, expectedArtifactResults[0].ArtifactID, ts.DB)
	require.Nil(ts.T(), err)
	// All artifact_results for the same artifact when created with seedArtifactResult. 
	require.Equal(ts.T(), 3, len(actualArtifactResults))
	requireDeepEqual(ts.T(), expectedArtifactResults, actualArtifactResults)
}

func (ts *TestSuite) TestArtifactResult_GetByArtifactAndWorkflow() {
	expectedArtifactResults, artifact, _, workflow  := ts.seedArtifactResult(3)

	actualArtifactResults, err := ts.artifact_result.GetByArtifactAndWorkflow(ts.ctx, workflow.ID, artifact.Name, ts.DB)

	require.Nil(ts.T(), err)
	// All artifact_results for the same artifact when created with seedArtifactResult. 
	require.Equal(ts.T(), 3, len(actualArtifactResults))
	requireDeepEqual(ts.T(), expectedArtifactResults, actualArtifactResults)
}

func (ts *TestSuite) TestArtifactResult_GetByArtifactAndDAGResult() {
	expectedArtifactResults, _, dag, _  := ts.seedArtifactResult(3)

	actualArtifactResults, err := ts.artifact_result.GetByArtifactAndDAGResult(ts.ctx, dag.ID, expectedArtifactResults[0].ID, ts.DB)

	require.Nil(ts.T(), err)
	// All artifact_results for the same artifact when created with seedArtifactResult. 
	require.Equal(ts.T(), 3, len(actualArtifactResults))
	requireDeepEqual(ts.T(), expectedArtifactResults, actualArtifactResults)
}

func (ts *TestSuite) TestArtifactResult_GetByDAGResults() {
	seedA = 3
	expectedArtifactResultsA, _, dagA, _  := ts.seedArtifactResult(seedA)
	seedB = 3
	// Generate artifact results for different DAG
	expectedArtifactResultsB, _, dagB, _  := ts.seedArtifactResult(seedB)

	actualArtifactResultsA, err := ts.artifact_result.GetByDAGResults(ts.ctx, []uuid.UUID{dagA.ID}, ts.DB)
	require.Nil(ts.T(), err)
	require.Equal(ts.T(), seedA, len(actualArtifactResultsA))
	requireDeepEqual(ts.T(), expectedArtifactResultsA, actualArtifactResultsA)

	actualArtifactResultsBoth, err := ts.artifact_result.GetByDAGResults(ts.ctx, []uuid.UUID{dagA.ID, dagB.ID}, ts.DB)
	require.Nil(ts.T(), err)
	require.Equal(ts.T(), seedA + seedB, len(actualArtifactResultsBoth))
	requireDeepEqual(ts.T(), append(expectedArtifactResultsA, expectedArtifactResultsB...), actualArtifactResultsBoth)
}

func (ts *TestSuite) TestArtifactResult_Create() {
	expectedArtifactResult := &models.ArtifactResult{
		DAGResultID: uuid.New(),
		ArtifactID: uuid.New(),
		ContentPath: randString(10),
	}

	actualArtifactResult, err := ts.artifact_result.Create(ts.ctx, expectedArtifactResult.DAGResultID, expectedArtifactResult.ArtifactID, expectedArtifactResult.ContentPath, ts.DB)
	require.Nil(ts.T(), err)

	require.NotEqual(ts.T(), uuid.Nil, actualArtifactResult.ID)

	expectedArtifactResult.ID = actualArtifactResult.ID

	requireDeepEqual(ts.T(), expectedArtifactResult, actualArtifactResult)
}

func (ts *TestSuite) TestArtifactResult_CreateWithExecStateAndMetadata() {
	schema := make([]map[string]string)
	schema[randString(10)] := randString(10)

	systemMetrics := make([]map[string]string)
	systemMetrics[randString(10)] := randString(10)

	expectedArtifactResult := &models.ArtifactResult{
		DAGResultID: uuid.New(),
		ArtifactID: uuid.New(),
		ContentPath: randString(10),
		ExecState: &shared.ExecutionState{
			UserLogs: &shared.Logs{
				Stdout:randString(10),
				StdErr:randString(10),
			},
			Status: shared.CanceledExecutionStatus,
		},
		Metadata: &shared.ArtifactResultMetadata{
			Schema: schema,
			SystemMetrics: systemMetrics,
			SerializationType: shared.StringSerialization,
			ArtifactType: shared.UntypedArtifact,
		},
	}

	actualArtifactResult, err := ts.artifact_result.CreateWithExecStateAndMetadata(ts.ctx, expectedArtifactResult.DAGResultID, expectedArtifactResult.ArtifactID, expectedArtifactResult.ContentPath, expectedArtifactResult.ExecState, expectedArtifactResult.Metadata, ts.DB)
	require.Nil(ts.T(), err)

	require.NotEqual(ts.T(), uuid.Nil, actualArtifactResult.ID)

	expectedArtifactResult.ID = actualArtifactResult.ID

	requireDeepEqual(ts.T(), expectedArtifactResult, actualArtifactResult)
}

func (ts *TestSuite) TestArtifactResult_Delete() {
	artifactResults, _, _, _ := ts.seedArtifactResult(1)

	err := ts.artifact_result.Delete(ts.ctx, artifactResults[0].ID, ts.DB)
	require.Nil(ts.T(), err)
}

func (ts *TestSuite) TestArtifactResult_DeleteBatch() {	
	artifactResults, _, _, _ := ts.seedArtifactResult(3)
	IDs := []uuid.UUID{artifactResults[0].ID, artifactResults[1].ID, artifactResults[2].ID}

	err := ts.dag.DeleteBatch(ts.ctx, IDs, ts.DB)
	require.Nil(ts.T(), err)
}

func (ts *TestSuite) TestArtifactResult_Update() {	
	artifactResults, _, _, _ := ts.seedArtifactResult(1)
	expectedArtifactResult := &artifactResults[0]

	contentPath := randString(10)
	execState := &shared.ExecutionState{
		UserLogs: &shared.Logs{
			Stdout:randString(10),
			StdErr:randString(10),
		},
		Status: shared.UnknownExecutionStatus,
	}
	metadata := &shared.ArtifactResultMetadata{
		Schema: schema,
		SystemMetrics: systemMetrics,
		SerializationType: shared.StringSerialization,
		ArtifactType: shared.JsonArtifact,
	}

	changes := map[string]interface{}{
		models.ArtifactResultContentPath: contentPath,
		models.ArtifactResultExecState: execState,
		models.ArtifactResultMetadata: metadata,
	}

	actualArtifactResult, err := ts.artifact_result.Update(ts.ctx, expectedArtifactResult.ID, changes, ts.DB)
	require.Nil(ts.T(), err)

	expectedArtifactResult.ContentPath = contentPath
	expectedArtifactResult.ExecState = execState
	expectedArtifactResult.Metadata = metadata

	requireDeepEqual(ts.T(), expectedArtifactResult, actualArtifactResult)
}
