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
	requireDeepEqualIntegrations(ts.T(), expectedIntegrations, actualIntegrations)
}

func (ts *TestSuite) TestIntegration_GetByConfigField() {
	integrations := ts.seedIntegration(3)

	for _, expectedIntegration := range integrations {
		// Because config is a random key-value string pair, assume no duplicates.
		for key, value := range expectedIntegration.config {
			actualIntegrations, err := ts.integration.GetByConfigField(ts.ctx, key, value, ts.DB)
			require.Equal(len(actualIntegrations), 1)
			actualIntegration := actualIntegrations[0]
			integrationValue, ok := actualIntegration.config[key]
			require.True(ok)
			require.Equal(value, integrationValue)
			requireDeepEqual(ts.T(), expectedIntegration, actualIntegration)
		}
	}
}

func (ts *TestSuite) TestIntegration_GetByNameAndUser() {
	integrations := ts.seedIntegration(3)

	actualIntegrations, err := ts.user.GetByNameAndUser(ts.ctx, expectedIntegrations[0].Name, expectedIntegrations[0].UserID, expectedIntegrations[0].OrgID, ts.DB)

	require.Nil(ts.T(), err)
	require.Equal(len(actualIntegrations), 3)
	requireDeepEqual(ts.T(), expectedIntegration, actualIntegrations)
}

func (ts *TestSuite) TestIntegration_GetByOrg() {
	expectedIntegrations := ts.seedIntegration(3)

	actualIntegrations, err := ts.user.GetByOrg(ts.ctx, expectedIntegrations[0].OrgID, ts.DB)

	require.Nil(ts.T(), err)
	require.Equal(len(actualIntegrations), 3)
	requireDeepEqualIntegrations(ts.T(), expectedIntegrations, actualIntegrations)
}

func (ts *TestSuite) TestIntegration_GetByServiceAndUser() {
	expectedIntegrations := ts.seedIntegration(3)

	actualIntegrations, err := ts.user.GetByServiceAndUser(ts.ctx, expectedIntegrations[0].Service, expectedIntegrations[0].UserID, ts.DB)

	require.Nil(ts.T(), err)
	require.Equal(len(actualIntegrations), 3)
	requireDeepEqualIntegrations(ts.T(), expectedIntegrations, actualIntegrations)
}

func (ts *TestSuite) TestIntegration_GetByUser() {
	expectedIntegrations := ts.seedIntegration(3)

	actualIntegrations, err := ts.user.GetByUser(ts.ctx, expectedIntegrations[0].OrgID, expectedIntegrations[0].UserID, ts.DB)

	require.Nil(ts.T(), err)
	require.Equal(ts.T(), len(actualIntegrations), 3)
	requireDeepEqualIntegrations(ts.T(), expectedIntegrations, actualIntegrations)
}

func (ts *TestSuite) TestIntegration_ValidateOwnership() {
	integrations := ts.seedIntegration(1)
	expectedIntegration := integrations[0]

	valid, err := ts.user.ValidateOwnership(ts.ctx, expectedIntegration.ID, expectedIntegration.OrgID, expectedIntegration.UserID, ts.DB)

	require.Nil(ts.T(), err)
	require.True(ts.T(), valid)
}

func (ts *TestSuite) TestIntegration_Create() {
	name := randString(10)
	config := {
		randString(10): randString(10),
	}
	valid := true

	expectedIntegration := &models.Integration{
		OrgID: testOrgID,
		Service: testIntegrationService,
		Name: name,
		Config: config,
		Validated: valid,
	}

	actualIntegration, err := ts.integration.Create(ts.ctx, expectedIntegration.OrgID, expectedIntegration.Service, expectedIntegration.Name, expectedIntegration.Config, expectedIntegration.Validated, ts.DB)
	require.Nil(ts.T(), err)

	require.NotEqual(ts.T(), uuid.Nil, actualIntegration.ID)

	expectedIntegration.ID = actualIntegration.ID
	expectedIntegration.CreatedAt = actualIntegration.CreatedAt
	requireDeepEqual(ts.T(), expectedIntegration, actualIntegration)
}

func (ts *TestSuite) TestIntegration_CreateForUser() {
	userID := uuid.New()
	name := randString(10)
	config := {
		randString(10): randString(10),
	}
	valid := true

	expectedIntegration := &models.Integration{
		UserID: userID,
		OrgID: testOrgID,
		Service: testIntegrationService,
		Name: name,
		Config: config,
		Validated: valid,
	}

	actualIntegration, err := ts.integration.Create(ts.ctx, expectedIntegration.OrgID, expectedIntegration.UserID, expectedIntegration.Service, expectedIntegration.Name, expectedIntegration.Config, expectedIntegration.Validated, ts.DB)
	require.Nil(ts.T(), err)

	require.NotEqual(ts.T(), uuid.Nil, actualIntegration.ID)

	expectedIntegration.ID = actualIntegration.ID
	expectedIntegration.CreatedAt = actualIntegration.CreatedAt
	requireDeepEqual(ts.T(), expectedIntegration, actualIntegration)
}

func (ts *TestSuite) TestIntegration_Delete() {	
	integrations := ts.seedIntegration(1)
	integration := integrations[0]

	err := ts.integration.Delete(ts.ctx, integration.ID, ts.DB)
	require.Nil(ts.T(), err)
}

func (ts *TestSuite) TestIntegration_Update() {	
	integrations := ts.seedIntegration(1)
	integration := integrations[0]

	name := randString(10)
	config := {
		randString(10): randString(10),
	}

	changes := map[string]interface{}{
		models.IntegrationName: name,
		models.IntegrationConfig: config,
	}

	newIntegration, err := ts.integration.Update(ts.ctx, integration.ID, changes, ts.DB)
	require.Nil(ts.T(), err)

	requireDeepEqual(ts.T(), name, newIntegration.Name)
	requireDeepEqual(ts.T(), config, newIntegration.Description)
}