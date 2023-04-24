package tests

import (
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/models/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func (ts *TestSuite) TestIntegration_Get() {
	integrations := ts.seedIntegration(1)
	expectedIntegration := &integrations[0]

	actualIntegration, err := ts.integration.Get(ts.ctx, expectedIntegration.ID, ts.DB)
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
		for key, value := range expectedIntegration.Config {
			actualIntegrations, err := ts.integration.GetByConfigField(ts.ctx, key, value, ts.DB)
			require.Nil(ts.T(), err)
			require.Equal(ts.T(), 1, len(actualIntegrations))
			actualIntegration := actualIntegrations[0]
			integrationValue, ok := actualIntegration.Config[key]
			require.True(ts.T(), ok)
			require.Equal(ts.T(), value, integrationValue)
			requireDeepEqual(ts.T(), expectedIntegration, actualIntegration)
		}
	}
}

func (ts *TestSuite) TestIntegration_GetByNameAndUser() {
	expectedIntegrations := ts.seedIntegration(1)
	expectedIntegration := expectedIntegrations[0]

	actualIntegration, err := ts.integration.GetByNameAndUser(
		ts.ctx,
		expectedIntegrations[0].Name,
		expectedIntegrations[0].UserID.UUID,
		expectedIntegrations[0].OrgID,
		ts.DB,
	)

	require.Nil(ts.T(), err)
	requireDeepEqual(ts.T(), expectedIntegration, *actualIntegration)
}

func (ts *TestSuite) TestIntegration_GetByOrg() {
	_ = ts.seedIntegration(3)

	actualIntegrations, err := ts.integration.GetByOrg(ts.ctx, testOrgID, ts.DB)

	require.Nil(ts.T(), err)
	require.Equal(ts.T(), 0, len(actualIntegrations))
}

func (ts *TestSuite) TestIntegration_GetByServiceAndUser() {
	expectedIntegrations := ts.seedIntegration(3)

	actualIntegrations, err := ts.integration.GetByServiceAndUser(ts.ctx, expectedIntegrations[0].Service, expectedIntegrations[0].UserID.UUID, ts.DB)

	require.Nil(ts.T(), err)
	require.Equal(ts.T(), 3, len(actualIntegrations))
	requireDeepEqualIntegrations(ts.T(), expectedIntegrations, actualIntegrations)
}

func (ts *TestSuite) TestIntegration_GetByUser() {
	expectedIntegrations := ts.seedIntegration(3)

	actualIntegrations, err := ts.integration.GetByUser(ts.ctx, expectedIntegrations[0].OrgID, expectedIntegrations[0].UserID.UUID, ts.DB)

	require.Nil(ts.T(), err)
	require.Equal(ts.T(), 3, len(actualIntegrations))
	requireDeepEqualIntegrations(ts.T(), expectedIntegrations, actualIntegrations)
}

func (ts *TestSuite) TestIntegration_ValidateOwnership() {
	integrations := ts.seedIntegration(1)
	expectedIntegration := integrations[0]

	valid, err := ts.integration.ValidateOwnership(ts.ctx, expectedIntegration.ID, expectedIntegration.OrgID, expectedIntegration.UserID.UUID, ts.DB)

	require.Nil(ts.T(), err)
	require.True(ts.T(), valid)
}

func (ts *TestSuite) TestIntegration_Create() {
	name := randString(10)
	config := make(shared.IntegrationConfig)
	config[randString(10)] = randString(10)

	expectedIntegration := &models.Integration{
		OrgID: testOrgID,
		UserID: utils.NullUUID{
			IsNull: true,
		},
		Service: testIntegrationService,
		Name:    name,
		Config:  config,
	}

	actualIntegration, err := ts.integration.Create(ts.ctx, expectedIntegration.OrgID, expectedIntegration.Service, expectedIntegration.Name, &expectedIntegration.Config, ts.DB)
	require.Nil(ts.T(), err)

	require.NotEqual(ts.T(), uuid.Nil, actualIntegration.ID)

	expectedIntegration.ID = actualIntegration.ID
	expectedIntegration.CreatedAt = actualIntegration.CreatedAt
	requireDeepEqual(ts.T(), expectedIntegration, actualIntegration)
}

func (ts *TestSuite) TestIntegration_CreateForUser() {
	userID := utils.NullUUID{
		UUID:   uuid.New(),
		IsNull: false,
	}
	name := randString(10)
	config := make(shared.IntegrationConfig)
	config[randString(10)] = randString(10)

	expectedIntegration := &models.Integration{
		UserID:  userID,
		OrgID:   testOrgID,
		Service: testIntegrationService,
		Name:    name,
		Config:  config,
	}

	actualIntegration, err := ts.integration.CreateForUser(
		ts.ctx,
		expectedIntegration.OrgID,
		expectedIntegration.UserID.UUID,
		expectedIntegration.Service,
		expectedIntegration.Name,
		&expectedIntegration.Config,
		ts.DB,
	)
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
	config := make(shared.IntegrationConfig)
	config[randString(10)] = randString(10)

	changes := map[string]interface{}{
		models.IntegrationName:   name,
		models.IntegrationConfig: &config,
	}

	newIntegration, err := ts.integration.Update(ts.ctx, integration.ID, changes, ts.DB)
	require.Nil(ts.T(), err)

	requireDeepEqual(ts.T(), name, newIntegration.Name)
	requireDeepEqual(ts.T(), config, newIntegration.Config)
}
