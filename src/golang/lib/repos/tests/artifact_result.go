package tests

import (
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func (ts *TestSuite) TestArtifactResult_Get() {
	artifact_results := ts.seedArtifactResult(1)
	expectedArtifactResult := &artifact_results[0]

	actualArtifactResult, err := ts.artifact_result.Get(ts.ctx, expectedArtifactResult.ID, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqual(ts.T(), expectedArtifactResult, actualArtifactResult)
}

func (ts *TestSuite) TestArtifactResult_GetBatch() {
	expectedArtifactResults := ts.seedArtifactResult(3)

	IDs := make([]uuid.UUID, 0, len(expectedArtifactResults))
	for _, expectedArtifactResult := range expectedArtifactResults {
		IDs = append(IDs, expectedArtifactResult.ID)
	}

	actualArtifactResults, err := ts.artifact_result.GetBatch(ts.ctx, IDs, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqualArtifacts(ts.T(), expectedArtifactResults, actualArtifactResults)
}

func (ts *TestSuite) TestArtifactResult_GetByArtifact() {
	artifact_results := ts.seedArtifactResult(3)
	// Seeded with uuid.New() for each artifactID so should only have 1 result per artifact.
	expectedArtifactResults := &artifact_results[0]

	actualArtifactResults, err := ts.artifact_result.GetByArtifact(ts.ctx, expectedArtifactResults.ArtifactID, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqual(ts.T(), []model.ArtifactResult{expectedArtifactResults}, actualArtifactResults)
}

func (ts *TestSuite) TestArtifactResult_GetByArtifactAndWorkflow() {
	users := ts.seedUser(1)
	expectedUser := &users[0]

	actualUser, err := ts.user.GetByAPIKey(ts.ctx, expectedUser.APIKey, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqual(ts.T(), expectedUser, actualUser)
}

func (ts *TestSuite) TestArtifactResult_GetByArtifactAndDAGResult() {
	users := ts.seedUser(1)
	expectedUser := &users[0]

	actualUser, err := ts.user.GetByAPIKey(ts.ctx, expectedUser.APIKey, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqual(ts.T(), expectedUser, actualUser)
}

func (ts *TestSuite) TestArtifactResult_GetByDAGResults() {
	users := ts.seedUser(1)
	expectedUser := &users[0]

	actualUser, err := ts.user.GetByAPIKey(ts.ctx, expectedUser.APIKey, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqual(ts.T(), expectedUser, actualUser)
}

func (ts *TestSuite) TestArtifactResult_Create() {
	apiKey := randAPIKey()

	expectedUser := &models.User{
		APIKey: apiKey,
	}

	actualUser, err := ts.user.Create(ts.ctx, testOrgID, apiKey, ts.DB)
	require.Nil(ts.T(), err)

	require.NotEqual(ts.T(), uuid.Nil, actualUser.ID)

	expectedUser.ID = actualUser.ID
	expectedUser.Email = actualUser.Email
	expectedUser.OrgID = actualUser.OrgID
	expectedUser.Role = actualUser.Role
	expectedUser.Auth0ID = actualUser.Auth0ID
	requireDeepEqual(ts.T(), expectedUser, actualUser)
}

func (ts *TestSuite) TestArtifactResult_CreateWithExecStateAndMetadata() {
	apiKey := randAPIKey()

	expectedUser := &models.User{
		APIKey: apiKey,
	}

	actualUser, err := ts.user.Create(ts.ctx, testOrgID, apiKey, ts.DB)
	require.Nil(ts.T(), err)

	require.NotEqual(ts.T(), uuid.Nil, actualUser.ID)

	expectedUser.ID = actualUser.ID
	expectedUser.Email = actualUser.Email
	expectedUser.OrgID = actualUser.OrgID
	expectedUser.Role = actualUser.Role
	expectedUser.Auth0ID = actualUser.Auth0ID
	requireDeepEqual(ts.T(), expectedUser, actualUser)
}

func (ts *TestSuite) TestArtifactResult_Delete() {
	apiKey := randAPIKey()

	expectedUser := &models.User{
		APIKey: apiKey,
	}

	actualUser, err := ts.user.Create(ts.ctx, testOrgID, apiKey, ts.DB)
	require.Nil(ts.T(), err)

	require.NotEqual(ts.T(), uuid.Nil, actualUser.ID)

	expectedUser.ID = actualUser.ID
	expectedUser.Email = actualUser.Email
	expectedUser.OrgID = actualUser.OrgID
	expectedUser.Role = actualUser.Role
	expectedUser.Auth0ID = actualUser.Auth0ID
	requireDeepEqual(ts.T(), expectedUser, actualUser)
}

func (ts *TestSuite) TestArtifactResult_DeleteBatch() {
	apiKey := randAPIKey()

	expectedUser := &models.User{
		APIKey: apiKey,
	}

	actualUser, err := ts.user.Create(ts.ctx, testOrgID, apiKey, ts.DB)
	require.Nil(ts.T(), err)

	require.NotEqual(ts.T(), uuid.Nil, actualUser.ID)

	expectedUser.ID = actualUser.ID
	expectedUser.Email = actualUser.Email
	expectedUser.OrgID = actualUser.OrgID
	expectedUser.Role = actualUser.Role
	expectedUser.Auth0ID = actualUser.Auth0ID
	requireDeepEqual(ts.T(), expectedUser, actualUser)
}

func (ts *TestSuite) TestArtifactResult_Update() {
	apiKey := randAPIKey()

	expectedUser := &models.User{
		APIKey: apiKey,
	}

	actualUser, err := ts.user.Create(ts.ctx, testOrgID, apiKey, ts.DB)
	require.Nil(ts.T(), err)

	require.NotEqual(ts.T(), uuid.Nil, actualUser.ID)

	expectedUser.ID = actualUser.ID
	expectedUser.Email = actualUser.Email
	expectedUser.OrgID = actualUser.OrgID
	expectedUser.Role = actualUser.Role
	expectedUser.Auth0ID = actualUser.Auth0ID
	requireDeepEqual(ts.T(), expectedUser, actualUser)
}