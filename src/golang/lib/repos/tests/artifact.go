package tests

import (
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func (ts *TestSuite) TestArtifact_Exists() {
	artifacts := ts.seedArtifact(1)
	expectedArtifact := artifacts[0]

	shouldExist, shouldExistErr := ts.artifact.Exists(ts.ctx, expectedArtifact.ID, ts.DB)
	require.Nil(ts.T(), shouldExistErr)
	require.True(ts.T(), shouldExist)

	shoudlNotExist, shoudlNotExistErr := ts.artifact.Exists(ts.ctx, uuid.New(), ts.DB)
	require.Nil(ts.T(), shoudlNotExistErr)
	require.False(ts.T(), shoudlNotExist)
}

func (ts *TestSuite) TestArtifact_Get() {	
	artifacts := ts.seedArtifact(1)
	expectedArtifact := artifacts[0]

	actualArtifact, err := ts.artifact.Get(ts.ctx, expectedArtifact.ID, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqual(ts.T(), expectedArtifact, *actualArtifact)
}

func (ts *TestSuite) TestArtifact_GetBatch() {
	expectedArtifacts := ts.seedArtifact(3)

	IDs := make([]uuid.UUID, 0, len(expectedArtifacts))
	for _, expectedArtifact := range expectedArtifacts {
		IDs = append(IDs, expectedArtifact.ID)
	}

	actualArtifact, err := ts.artifact.GetBatch(ts.ctx, IDs, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqualArtifacts(ts.T(), expectedArtifacts, actualArtifact)
}

func (ts *TestSuite) TestArtifact_GetByDAG() {
	expectedArtifact, dag, _, _ := ts.seedArtifactInWorkflow()

	actualArtifact, err := ts.artifact.GetByDAG(ts.ctx, dag.ID, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqual(ts.T(), []models.Artifact{expectedArtifact}, actualArtifact)
}

func (ts *TestSuite) TestArtifact_ValidateOrg() {
	expectedArtifact, _, _, user := ts.seedArtifactInWorkflow()

	createdByOrg, createdByOrgErr := ts.artifact.ValidateOrg(ts.ctx, expectedArtifact.ID, user.OrgID, ts.DB)
	require.Nil(ts.T(), createdByOrgErr)
	require.True(ts.T(), createdByOrg)

	notCreatedByOrg, notCreatedByOrgErr := ts.artifact.ValidateOrg(ts.ctx, expectedArtifact.ID, randString(15), ts.DB)
	require.Nil(ts.T(), notCreatedByOrgErr)
	require.False(ts.T(), notCreatedByOrg)
}

func (ts *TestSuite) TestArtifact_Create() {
	name := randString(10)
	description := randString(15)
	artifactType := randArtifactType()

	expectedArtifact := &models.Artifact{
		Name: name,
		Description: description,
		Type: artifactType,
	}

	actualArtifact, err := ts.artifact.Create(ts.ctx, name, description, artifactType, ts.DB)
	require.Nil(ts.T(), err)

	require.NotEqual(ts.T(), uuid.Nil, actualArtifact.ID)

	expectedArtifact.ID = actualArtifact.ID
	requireDeepEqual(ts.T(), expectedArtifact, actualArtifact)
}

func (ts *TestSuite) TestArtifact_Update() {
	artifacts := ts.seedArtifact(1)
	artifact := artifacts[0]

	name := randString(10)
	description := randString(15)
	artifactType := randArtifactType()

	changes := map[string]interface{}{
		models.ArtifactName: name,
		models.ArtifactDescription: description,
		models.ArtifactType: artifactType,
	}

	newArtifact, err := ts.artifact.Update(ts.ctx, artifact.ID, changes, ts.DB)
	require.Nil(ts.T(), err)

	requireDeepEqual(ts.T(), name, newArtifact.Name)
	requireDeepEqual(ts.T(), description, newArtifact.Description)
	requireDeepEqual(ts.T(), artifactType, newArtifact.Type)
}

func (ts *TestSuite) TestArtifact_Delete() {
	artifacts := ts.seedArtifact(1)
	artifact := artifacts[0]

	err := ts.artifact.Delete(ts.ctx, artifact.ID, ts.DB)
	require.Nil(ts.T(), err)
}

func (ts *TestSuite) TestArtifact_DeleteBatch() {
	artifacts := ts.seedArtifact(3)

	IDs := make([]uuid.UUID, 0, len(artifacts))
	for _, artifact := range artifacts {
		IDs = append(IDs, artifact.ID)
	}
	
	err := ts.artifact.DeleteBatch(ts.ctx, IDs, ts.DB)
	require.Nil(ts.T(), err)
}
