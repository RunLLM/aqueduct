package tests

import (
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/models/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func (ts *TestSuite) TestResource_Get() {
	resources := ts.seedResource(1)
	expectedResource := &resources[0]

	actualResource, err := ts.resource.Get(ts.ctx, expectedResource.ID, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqual(ts.T(), expectedResource, actualResource)
}

func (ts *TestSuite) TestResource_GetBatch() {
	expectedResources := ts.seedResource(3)

	IDs := make([]uuid.UUID, 0, len(expectedResources))
	for _, expectedResource := range expectedResources {
		IDs = append(IDs, expectedResource.ID)
	}

	actualResources, err := ts.resource.GetBatch(ts.ctx, IDs, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqualResources(ts.T(), expectedResources, actualResources)
}

func (ts *TestSuite) TestResource_GetByConfigField() {
	resources := ts.seedResource(3)

	for _, expectedResource := range resources {
		// Because config is a random key-value string pair, assume no duplicates.
		for key, value := range expectedResource.Config {
			actualResources, err := ts.resource.GetByConfigField(ts.ctx, key, value, ts.DB)
			require.Nil(ts.T(), err)
			require.Equal(ts.T(), 1, len(actualResources))
			actualResource := actualResources[0]
			resourceValue, ok := actualResource.Config[key]
			require.True(ts.T(), ok)
			require.Equal(ts.T(), value, resourceValue)
			requireDeepEqual(ts.T(), expectedResource, actualResource)
		}
	}
}

func (ts *TestSuite) TestResource_GetByNameAndUser() {
	expectedResources := ts.seedResource(1)
	expectedResource := expectedResources[0]

	actualResource, err := ts.resource.GetByNameAndUser(
		ts.ctx,
		expectedResources[0].Name,
		expectedResources[0].UserID.UUID,
		expectedResources[0].OrgID,
		ts.DB,
	)

	require.Nil(ts.T(), err)
	requireDeepEqual(ts.T(), expectedResource, *actualResource)
}

func (ts *TestSuite) TestResource_GetByOrg() {
	_ = ts.seedResource(3)

	actualResources, err := ts.resource.GetByOrg(ts.ctx, testOrgID, ts.DB)

	require.Nil(ts.T(), err)
	require.Equal(ts.T(), 0, len(actualResources))
}

func (ts *TestSuite) TestResource_GetByServiceAndUser() {
	expectedResources := ts.seedResource(3)

	actualResources, err := ts.resource.GetByServiceAndUser(ts.ctx, expectedResources[0].Service, expectedResources[0].UserID.UUID, ts.DB)

	require.Nil(ts.T(), err)
	require.Equal(ts.T(), 3, len(actualResources))
	requireDeepEqualResources(ts.T(), expectedResources, actualResources)
}

func (ts *TestSuite) TestResource_GetByUser() {
	expectedResources := ts.seedResource(3)

	actualResources, err := ts.resource.GetByUser(ts.ctx, expectedResources[0].OrgID, expectedResources[0].UserID.UUID, ts.DB)

	require.Nil(ts.T(), err)
	require.Equal(ts.T(), 3, len(actualResources))
	requireDeepEqualResources(ts.T(), expectedResources, actualResources)
}

func (ts *TestSuite) TestResource_ValidateOwnership() {
	resources := ts.seedResource(1)
	expectedResource := resources[0]

	valid, err := ts.resource.ValidateOwnership(ts.ctx, expectedResource.ID, expectedResource.OrgID, expectedResource.UserID.UUID, ts.DB)

	require.Nil(ts.T(), err)
	require.True(ts.T(), valid)
}

func (ts *TestSuite) TestResource_Create() {
	name := randString(10)
	config := make(shared.ResourceConfig)
	config[randString(10)] = randString(10)

	expectedResource := &models.Resource{
		OrgID: testOrgID,
		UserID: utils.NullUUID{
			IsNull: true,
		},
		Service: testResourceService,
		Name:    name,
		Config:  config,
	}

	actualResource, err := ts.resource.Create(ts.ctx, expectedResource.OrgID, expectedResource.Service, expectedResource.Name, &expectedResource.Config, ts.DB)
	require.Nil(ts.T(), err)

	require.NotEqual(ts.T(), uuid.Nil, actualResource.ID)

	expectedResource.ID = actualResource.ID
	expectedResource.CreatedAt = actualResource.CreatedAt
	requireDeepEqual(ts.T(), expectedResource, actualResource)
}

func (ts *TestSuite) TestResource_CreateForUser() {
	userID := utils.NullUUID{
		UUID:   uuid.New(),
		IsNull: false,
	}
	name := randString(10)
	config := make(shared.ResourceConfig)
	config[randString(10)] = randString(10)

	expectedResource := &models.Resource{
		UserID:  userID,
		OrgID:   testOrgID,
		Service: testResourceService,
		Name:    name,
		Config:  config,
	}

	actualResource, err := ts.resource.CreateForUser(
		ts.ctx,
		expectedResource.OrgID,
		expectedResource.UserID.UUID,
		expectedResource.Service,
		expectedResource.Name,
		&expectedResource.Config,
		ts.DB,
	)
	require.Nil(ts.T(), err)

	require.NotEqual(ts.T(), uuid.Nil, actualResource.ID)

	expectedResource.ID = actualResource.ID
	expectedResource.CreatedAt = actualResource.CreatedAt
	requireDeepEqual(ts.T(), expectedResource, actualResource)
}

func (ts *TestSuite) TestResource_Delete() {
	resources := ts.seedResource(1)
	resource := resources[0]

	err := ts.resource.Delete(ts.ctx, resource.ID, ts.DB)
	require.Nil(ts.T(), err)
}

func (ts *TestSuite) TestResource_Update() {
	resources := ts.seedResource(1)
	resource := resources[0]

	name := randString(10)
	config := make(shared.ResourceConfig)
	config[randString(10)] = randString(10)

	changes := map[string]interface{}{
		models.ResourceName:   name,
		models.ResourceConfig: &config,
	}

	newResource, err := ts.resource.Update(ts.ctx, resource.ID, changes, ts.DB)
	require.Nil(ts.T(), err)

	requireDeepEqual(ts.T(), name, newResource.Name)
	requireDeepEqual(ts.T(), config, newResource.Config)
}
