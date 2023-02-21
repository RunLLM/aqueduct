package tests

import (
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func (ts *TestSuite) TestExecutionEnvironment_Get() {
	executionEnvironments := ts.seedUnusedExecutionEnvironment(1)
	expectedExecutionEnvironment := &executionEnvironments[0]

	actualExecutionEnvironment, err := ts.executionEnvironment.Get(ts.ctx, expectedExecutionEnvironment.ID, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqual(ts.T(), expectedExecutionEnvironment, actualExecutionEnvironment)
}

func (ts *TestSuite) TestExecutionEnvironment_GetBatch() {
	expectedExecutionEnvironments := ts.seedUnusedExecutionEnvironment(3)

	IDs := make([]uuid.UUID, 0, len(expectedExecutionEnvironments))
	for _, expectedExecutionEnvironment := range expectedExecutionEnvironments {
		IDs = append(IDs, expectedExecutionEnvironment.ID)
	}

	actualExecutionEnvironments, err := ts.executionEnvironment.GetBatch(ts.ctx, IDs, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqualExecutionEnvironment(ts.T(), expectedExecutionEnvironments, actualExecutionEnvironments)
}

func (ts *TestSuite) TestExecutionEnvironment_GetActiveByHash() {
	executionEnvironments := ts.seedUnusedExecutionEnvironment(1)
	expectedExecutionEnvironment := &executionEnvironments[0]

	actualExecutionEnvironment, err := ts.executionEnvironment.GetActiveByHash(ts.ctx, expectedExecutionEnvironment.Hash, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqual(ts.T(), expectedExecutionEnvironment, actualExecutionEnvironment)
}

func (ts *TestSuite) TestExecutionEnvironment_GetActiveByOperatorBatch() {
	expectedExecutionEnvironments, operators := ts.seedUsedExecutionEnvironment(6)

	actualExecutionEnvironments, err := ts.executionEnvironment.GetActiveByOperatorBatch(
		ts.ctx,
		[]uuid.UUID{operators[0].ID, operators[2].ID, operators[4].ID},
		ts.DB,
	)

	require.Nil(ts.T(), err)
	require.Equal(ts.T(), 3, len(actualExecutionEnvironments))
	requireDeepEqualExecutionEnvironment(ts.T(),
		[]models.ExecutionEnvironment{
			expectedExecutionEnvironments[0],
			expectedExecutionEnvironments[2],
			expectedExecutionEnvironments[4],
		},
		[]models.ExecutionEnvironment{
			actualExecutionEnvironments[operators[0].ID],
			actualExecutionEnvironments[operators[2].ID],
			actualExecutionEnvironments[operators[4].ID],
		})
}

func (ts *TestSuite) TestExecutionEnvironment_GetUnused() {
	ts.seedUsedExecutionEnvironment(3)

	noNewExecutionEnvironments, err := ts.executionEnvironment.GetUnused(ts.ctx, ts.DB)
	require.Nil(ts.T(), err)
	require.Equal(ts.T(), 0, len(noNewExecutionEnvironments))

	expectedExecutionEnvironments := ts.seedUnusedExecutionEnvironment(3)

	actualExecutionEnvironments, err := ts.executionEnvironment.GetUnused(ts.ctx, ts.DB)
	require.Nil(ts.T(), err)
	require.Equal(ts.T(), 3, len(actualExecutionEnvironments))
	requireDeepEqualExecutionEnvironment(ts.T(), expectedExecutionEnvironments, actualExecutionEnvironments)
}

func (ts *TestSuite) TestExecutionEnvironment_Create() {
	spec := shared.ExecutionEnvironmentSpec{
		PythonVersion: randString(10),
		Dependencies:  []string{randString(10), randString(10), randString(10)},
	}
	hash := uuid.New()

	expectedExecutionEnvironment := &models.ExecutionEnvironment{
		Spec: spec,
		Hash: hash,
	}

	actualExecutionEnvironment, err := ts.executionEnvironment.Create(ts.ctx, &expectedExecutionEnvironment.Spec, expectedExecutionEnvironment.Hash, ts.DB)
	require.Nil(ts.T(), err)

	require.NotEqual(ts.T(), uuid.Nil, actualExecutionEnvironment.ID)

	expectedExecutionEnvironment.ID = actualExecutionEnvironment.ID

	requireDeepEqual(ts.T(), expectedExecutionEnvironment, actualExecutionEnvironment)
}

func (ts *TestSuite) TestExecutionEnvironment_Delete() {
	executionEnvironments := ts.seedUnusedExecutionEnvironment(1)

	err := ts.artifactResult.Delete(ts.ctx, executionEnvironments[0].ID, ts.DB)
	require.Nil(ts.T(), err)
}

func (ts *TestSuite) TestExecutionEnvironment_DeleteBatch() {
	executionEnvironments := ts.seedUnusedExecutionEnvironment(3)

	IDs := make([]uuid.UUID, 0, len(executionEnvironments))
	for _, executionEnvironment := range executionEnvironments {
		IDs = append(IDs, executionEnvironment.ID)
	}

	err := ts.dag.DeleteBatch(ts.ctx, IDs, ts.DB)
	require.Nil(ts.T(), err)
}

func (ts *TestSuite) TestExecutionEnvironment_Update() {
	executionEnvironments := ts.seedUnusedExecutionEnvironment(1)
	expectedExecutionEnvironment := &executionEnvironments[0]

	spec := shared.ExecutionEnvironmentSpec{
		PythonVersion: randString(10),
		Dependencies:  []string{randString(10), randString(10), randString(10)},
	}
	hash := uuid.New()

	changes := map[string]interface{}{
		models.ExecutionEnvironmentSpec: &spec,
		models.ExecutionEnvironmentHash: hash,
	}

	actualExecutionEnvironment, err := ts.executionEnvironment.Update(ts.ctx, expectedExecutionEnvironment.ID, changes, ts.DB)
	require.Nil(ts.T(), err)

	expectedExecutionEnvironment.Spec = spec
	expectedExecutionEnvironment.Hash = hash

	requireDeepEqual(ts.T(), expectedExecutionEnvironment, actualExecutionEnvironment)
}
