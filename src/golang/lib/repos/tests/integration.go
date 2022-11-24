package tests

import (
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func (ts *TestSuite) TestIntegration_Get() {
	integrations := ts.seedIntegration(1)
	expectedIntegration := &integrations[0]

	actualIntegration, err := ts.user.Get(ts.ctx, expectedUser.ID, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqual(ts.T(), expectedIntegration, actualIntegration)
}

func (ts *TestSuite) TestIntegration_GetBatch() {	
	expectedIntegrations := ts.seedIntegration(3)

	IDs := make([]uuid.UUID, 0, len(expectedIntegrations))
	for _, expectedIntegration := range expectedIntegrations {
		IDs = append(IDs, expectedIntegration.ID)
	}

	actualIntegrations, err := ts.integration.GetBatch(ts.ctx, IDs, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqualArtifacts(ts.T(), expectedIntegrations, actualIntegrations)
}

func (ts *TestSuite) TestIntegration_GetByConfigField() {
	users := ts.seedUser(1)
	expectedUser := &users[0]

	actualUser, err := ts.user.GetByAPIKey(ts.ctx, expectedUser.APIKey, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqual(ts.T(), expectedUser, actualUser)
}

func (ts *TestSuite) TestIntegration_GetByNameAndUser() {
	users := ts.seedUser(1)
	expectedUser := &users[0]

	actualUser, err := ts.user.GetByAPIKey(ts.ctx, expectedUser.APIKey, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqual(ts.T(), expectedUser, actualUser)
}

func (ts *TestSuite) TestIntegration_GetByOrg() {
	users := ts.seedUser(1)
	expectedUser := &users[0]

	actualUser, err := ts.user.GetByAPIKey(ts.ctx, expectedUser.APIKey, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqual(ts.T(), expectedUser, actualUser)
}

func (ts *TestSuite) TestIntegration_GetByServiceAndUser() {
	users := ts.seedUser(1)
	expectedUser := &users[0]

	actualUser, err := ts.user.GetByAPIKey(ts.ctx, expectedUser.APIKey, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqual(ts.T(), expectedUser, actualUser)
}

func (ts *TestSuite) TestIntegration_GetByUser() {
	users := ts.seedUser(1)
	expectedUser := &users[0]

	actualUser, err := ts.user.GetByAPIKey(ts.ctx, expectedUser.APIKey, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqual(ts.T(), expectedUser, actualUser)
}

func (ts *TestSuite) TestIntegration_ValidateOwnership() {
	users := ts.seedUser(1)
	expectedUser := &users[0]

	actualUser, err := ts.user.GetByAPIKey(ts.ctx, expectedUser.APIKey, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqual(ts.T(), expectedUser, actualUser)
}

func (ts *TestSuite) TestIntegration_Create() {
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

func (ts *TestSuite) TestIntegration_CreateForUser() {
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

func (ts *TestSuite) TestIntegration_Delete() {
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

func (ts *TestSuite) TestIntegration_Update() {
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