package tests

import (
	"context"
	"testing"

	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	"github.com/aqueducthq/aqueduct/lib/collections/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func seedIntegration(t *testing.T, count int) []integration.Integration {
	numUsers := 2

	users := seedUser(t, numUsers)
	require.Equal(t, numUsers, len(users))
	orgIds := randOrgIdsFromList(count, users)

	return seedIntegrationWithOrgIds(t, count, orgIds)
}

// seedIntegrationWithOrgIds populates the integration table with count integrations
// where organization_id is set to the values provided in orgIds.
func seedIntegrationWithOrgIds(t *testing.T, count int, orgIds []string) []integration.Integration {
	require.Len(t, orgIds, count)

	integrations := make([]integration.Integration, 0, count)

	for i := 0; i < count; i++ {
		testIntegration, err := writers.integrationWriter.CreateIntegration(
			context.Background(),
			orgIds[i],
			integration.Postgres,
			randString(10),
			(*utils.Config)(&map[string]string{
				"username": "user",
				"password": "pwd",
			}),
			true,
			db,
		)
		require.Nil(t, err)

		integrations = append(integrations, *testIntegration)
	}

	require.Len(t, integrations, count)

	return integrations
}

func TestCreateIntegration(t *testing.T) {
	defer resetDatabase(t)

	users := seedUser(t, 1)

	expectedIntegration := &integration.Integration{
		UserId: utils.NullUUID{
			IsNull: true,
		},
		OrganizationId: users[0].OrganizationId,
		Service:        integration.Snowflake,
		Name:           "test-integration",
		Config: utils.Config(map[string]string{
			"username": "user",
			"password": "pwd",
		}),
		Validated: true,
	}

	actualIntegration, err := writers.integrationWriter.CreateIntegration(
		context.Background(),
		expectedIntegration.OrganizationId,
		expectedIntegration.Service,
		expectedIntegration.Name,
		&expectedIntegration.Config,
		expectedIntegration.Validated,
		db,
	)
	require.Nil(t, err)
	require.NotEqual(t, uuid.Nil, actualIntegration.Id)

	expectedIntegration.Id = actualIntegration.Id
	expectedIntegration.CreatedAt = actualIntegration.CreatedAt

	requireDeepEqual(t, expectedIntegration, actualIntegration)
}

func TestCreateIntegrationForUser(t *testing.T) {
	defer resetDatabase(t)

	users := seedUser(t, 1)

	expectedIntegration := &integration.Integration{
		UserId: utils.NullUUID{
			IsNull: false,
			UUID:   users[0].Id,
		},
		OrganizationId: users[0].OrganizationId,
		Service:        integration.Snowflake,
		Name:           "test-integration",
		Config: utils.Config(map[string]string{
			"username": "user",
			"password": "pwd",
		}),
		Validated: true,
	}

	actualIntegration, err := writers.integrationWriter.CreateIntegrationForUser(
		context.Background(),
		expectedIntegration.OrganizationId,
		expectedIntegration.UserId.UUID,
		expectedIntegration.Service,
		expectedIntegration.Name,
		&expectedIntegration.Config,
		expectedIntegration.Validated,
		db,
	)
	require.Nil(t, err)
	require.NotEqual(t, uuid.Nil, actualIntegration.Id)

	expectedIntegration.Id = actualIntegration.Id
	expectedIntegration.CreatedAt = actualIntegration.CreatedAt

	requireDeepEqual(t, expectedIntegration, actualIntegration)
}

func TestGetIntegrationsByConfigField(t *testing.T) {
	defer resetDatabase(t)

	numIntegrations := 2

	otherIntegrations := seedIntegration(t, numIntegrations)

	// Create another integration with specific config
	testIntegration, err := writers.integrationWriter.CreateIntegration(
		context.Background(),
		otherIntegrations[0].OrganizationId,
		integration.BigQuery,
		randString(10),
		(*utils.Config)(&map[string]string{
			"username": "special-username",
			"password": "special-password",
		}),
		true,
		db,
	)
	require.Nil(t, err)

	actualIntegrations, err := readers.integrationReader.GetIntegrationsByConfigField(
		context.Background(),
		"username",
		"special-username",
		db,
	)
	require.Nil(t, err)
	require.Equal(t, 1, len(actualIntegrations))

	requireDeepEqual(t, *testIntegration, actualIntegrations[0])
}
