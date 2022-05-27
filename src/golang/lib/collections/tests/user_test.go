package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/aqueducthq/aqueduct/lib/collections/user"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

const (
	testOrganizationId = "123456"
)

func seedUser(t *testing.T, count int) []user.User {
	users := make([]user.User, 0, count)

	for i := 0; i < count; i++ {
		email := fmt.Sprintf("%s@aqueducthq.com", randString(10))
		orgId := testOrganizationId
		role := string(user.AdminRole)
		auth0Id := randString(10)

		testUser, err := writers.userWriter.CreateUser(
			context.Background(),
			email,
			orgId,
			role,
			auth0Id,
			db,
		)
		require.Nil(t, err)

		users = append(users, *testUser)
	}

	require.Len(t, users, count)

	return users
}

func TestCreateUser(t *testing.T) {
	defer resetDatabase(t)

	expectedUser := &user.User{
		Email:          "test@aqueducthq.com",
		OrganizationId: testOrganizationId,
		Role:           string(user.AdminRole),
		Auth0Id:        "auth123",
	}

	actualUser, err := writers.userWriter.CreateUser(
		context.Background(),
		expectedUser.Email,
		expectedUser.OrganizationId,
		expectedUser.Role,
		expectedUser.Auth0Id,
		db,
	)
	require.Nil(t, err)

	require.NotEqual(t, uuid.Nil, actualUser.Id)

	expectedUser.Id = actualUser.Id
	expectedUser.ApiKey = actualUser.ApiKey

	requireDeepEqual(t, expectedUser, actualUser)
}

func TestGetUser(t *testing.T) {
	defer resetDatabase(t)

	numUsers := 1

	users := seedUser(t, numUsers)

	expectedUser := users[0]
	actualUser, err := readers.userReader.GetUser(context.Background(), expectedUser.Id, db)
	require.Nil(t, err)
	requireDeepEqual(t, &expectedUser, actualUser)
}

func TestGetUsersInOrganization(t *testing.T) {
	defer resetDatabase(t)

	numUsers := 3

	users := seedUser(t, numUsers)

	actualUsers, err := readers.userReader.GetUsersInOrganization(context.Background(), testOrganizationId, db)
	require.Nil(t, err)

	for _, expectedUser := range users {
		found := false
		for _, actualUser := range actualUsers {
			if expectedUser.Id == actualUser.Id {
				found = true
				requireDeepEqual(t, &expectedUser, &actualUser)
			}
		}
		require.True(t, found, "Unable to find user %v", expectedUser)
	}
}

func TestDeleteUser(t *testing.T) {
	defer resetDatabase(t)

	numUsers := 1

	users := seedUser(t, numUsers)
	toDelete := users[0]

	err := writers.userWriter.DeleteUser(context.Background(), toDelete.Id, db)
	require.Nil(t, err)

	_, err = readers.userReader.GetUser(context.Background(), toDelete.Id, db)
	require.EqualError(t, err, database.ErrNoRows.Error())
}
