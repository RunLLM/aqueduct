package tests

import (
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func (ts *TestSuite) TestDAG_Get() {
	dags := ts.seedDAG(1)
	expectedDAG := dags[0]

	actualDAG, err := ts.dag.Get(ts.ctx, expectedDAG.ID, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqual(ts.T(), expectedDAG, *actualDAG)
}

func (ts *TestSuite) TestDAG_GetBatch() {
	expectedDAGs := ts.seedDAG(3)

	IDs := make([]uuid.UUID, 0, len(expectedDAGs))
	for _, expectedDAG := range expectedDAGs {
		IDs = append(IDs, expectedDAG.ID)
	}

	actualDAGs, err := ts.dag.GetBatch(ts.ctx, IDs, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqualDAGs(ts.T(), expectedDAGs, actualDAGs)
}

func (ts *TestSuite) TestDAG_GetByArtifactResultBatch() {
	// TODO: Complete after artifact result is refactored
}

func (ts *TestSuite) TestDAG_GetByDAGResult() {
	// TODO: Complete after DAG result is refactored
}

func (ts *TestSuite) TestDAG_GetByOperator() {
	// TODO: Complete after Operator is refactored
}

func (ts *TestSuite) TestDAG_GetByWorkflow() {
	workflows := ts.seedWorkflow(1)
	workflow := workflows[0]

	expectedDAGs := ts.seedDAGWithWorkflow(3, []uuid.UUID{workflow.ID, workflow.ID, workflow.ID})

	actualDAGs, err := ts.dag.GetByWorkflow(ts.ctx, workflow.ID, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqualDAGs(ts.T(), expectedDAGs, actualDAGs)
}

func (ts *TestSuite) TestDAG_GetLatestByWorkflow() {
	workflows := ts.seedWorkflow(1)
	workflow := workflows[0]

	dags := ts.seedDAGWithWorkflow(3, []uuid.UUID{workflow.ID, workflow.ID, workflow.ID})

	expectedDAG := dags[0]
	for i := 1; i < len(dags); i++ {
		if dags[i].CreatedAt.After(expectedDAG.CreatedAt) {
			expectedDAG = dags[i]
		}
	}

	actualDAG, err := ts.dag.GetLatestByWorkflow(ts.ctx, workflow.ID, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqual(ts.T(), expectedDAG, *actualDAG)
}

func (ts *TestSuite) TestDAG_Create() {
	workflows := ts.seedWorkflow(1)
	workflow := workflows[0]

	expectedDAG := &models.DAG{
		WorkflowID: workflow.ID,
		StorageConfig: shared.StorageConfig{
			Type: shared.S3StorageType,
			S3Config: &shared.S3Config{
				Region:             "us-east-2",
				Bucket:             "test",
				CredentialsPath:    "/home/users/.aws/credentials",
				CredentialsProfile: "default",
			},
		},
		EngineConfig: shared.EngineConfig{
			Type:           shared.AqueductEngineType,
			AqueductConfig: &shared.AqueductConfig{},
		},
	}

	actualDAG, err := ts.dag.Create(
		ts.ctx,
		expectedDAG.WorkflowID,
		&expectedDAG.StorageConfig,
		&expectedDAG.EngineConfig,
		ts.DB,
	)
	require.Nil(ts.T(), err)

	require.NotEqual(ts.T(), uuid.Nil, actualDAG.ID)

	expectedDAG.ID = actualDAG.ID
	expectedDAG.CreatedAt = actualDAG.CreatedAt
	requireDeepEqual(ts.T(), expectedDAG, actualDAG)
}

func (ts *TestSuite) TestDAG_Delete() {
	dags := ts.seedDAG(1)
	dag := dags[0]

	err := ts.dag.Delete(ts.ctx, dag.ID, ts.DB)
	require.Nil(ts.T(), err)
}

func (ts *TestSuite) TestDAG_DeleteBatch() {
	dags := ts.seedDAG(2)
	IDs := []uuid.UUID{dags[0].ID, dags[1].ID}

	err := ts.dag.DeleteBatch(ts.ctx, IDs, ts.DB)
	require.Nil(ts.T(), err)
}

func (ts *TestSuite) TestDAG_Update() {
	dags := ts.seedDAG(1)
	dag := dags[0]

	newStorageConfig := shared.StorageConfig{
		Type: shared.GCSStorageType,
		GCSConfig: &shared.GCSConfig{
			Bucket:                    "test-gcs",
			ServiceAccountCredentials: "credentials",
		},
	}

	changes := map[string]interface{}{
		models.DagStorageConfig: &newStorageConfig,
	}

	newDAG, err := ts.dag.Update(ts.ctx, dag.ID, changes, ts.DB)
	require.Nil(ts.T(), err)

	requireDeepEqual(ts.T(), newStorageConfig, newDAG.StorageConfig)
}
