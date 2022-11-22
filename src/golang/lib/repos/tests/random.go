package tests

import (
	"math/rand"

	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
)

// randArtifactType generates a random artifact type.
func randArtifactType() shared.ArtifactType {
	return shared.ArtifactTypes[rand.Intn(len(shared.ArtifactTypes))]
}

// randAPIKey generates a random API key.
func randAPIKey() string {
	return randString(60)
}

// randString generates a random string of length n.
func randString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))] // nolint:gosec
	}
	return string(b)
}

// sampleIDs randomly samples IDs n times with replacement.
func sampleIDs(n int, IDs []uuid.UUID) []uuid.UUID {
	polled := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		ID := IDs[rand.Intn(len(IDs))]
		polled = append(polled, ID)
	}
	return polled
}

// sampleUserIDs randomly samples users n times with replacement,
// and returns the ID of selected Users.
func sampleUserIDs(n int, users []models.User) []uuid.UUID {
	userIDs := make([]uuid.UUID, 0, len(users))
	for _, user := range users {
		userIDs = append(userIDs, user.ID)
	}
	return sampleIDs(n, userIDs)
}
