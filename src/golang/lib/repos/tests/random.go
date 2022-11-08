package tests

import (
	"math/rand"

	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/google/uuid"
)

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

// pollIDs polls IDs n times.
func pollIDs(n int, IDs []uuid.UUID) []uuid.UUID {
	polled := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		ID := IDs[rand.Intn(len(IDs))]
		polled = append(polled, ID)
	}
	return polled
}

// pollUserIDs polls users n times and returns the IDs of those users.
func pollUserIDs(n int, users []models.User) []uuid.UUID {
	userIDs := make([]uuid.UUID, 0, len(users))
	for _, user := range users {
		userIDs = append(userIDs, user.ID)
	}
	return pollIDs(n, userIDs)
}
