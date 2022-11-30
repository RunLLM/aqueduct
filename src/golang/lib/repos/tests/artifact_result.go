package tests

import (
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func (ts *TestSuite) TestArtifactResult_Get() {
	users := ts.seedUser(1)
	expectedUser := &users[0]

	actualUser, err := ts.user.GetByAPIKey(ts.ctx, expectedUser.APIKey, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqual(ts.T(), expectedUser, actualUser)
}

func (ts *TestSuite) TestArtifactResult_GetBatch() {
	users := ts.seedUser(1)
	expectedUser := &users[0]

	actualUser, err := ts.user.GetByAPIKey(ts.ctx, expectedUser.APIKey, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqual(ts.T(), expectedUser, actualUser)
}

func (ts *TestSuite) TestArtifactResult_GetByArtifact() {
	users := ts.seedUser(1)
	expectedUser := &users[0]

	actualUser, err := ts.user.GetByAPIKey(ts.ctx, expectedUser.APIKey, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqual(ts.T(), expectedUser, actualUser)
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