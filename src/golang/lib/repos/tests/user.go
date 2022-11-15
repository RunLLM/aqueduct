package tests

import (
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func (ts *TestSuite) TestUser_GetByAPIKey() {
	users := ts.seedUser(1)
	expectedUser := &users[0]

	actualUser, err := ts.user.GetByAPIKey(ts.ctx, expectedUser.APIKey, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqual(ts.T(), expectedUser, actualUser)
}

func (ts *TestSuite) TestUser_Create() {
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

func (ts *TestSuite) TestUser_ResetAPIKey() {
	users := ts.seedUser(1)
	user := users[0]

	updatedUser, err := ts.user.ResetAPIKey(ts.ctx, user.ID, ts.DB)
	require.Nil(ts.T(), err)

	require.NotEqual(ts.T(), user.APIKey, updatedUser.APIKey)
}
