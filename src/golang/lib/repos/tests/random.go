package tests

import (
	"math/rand"

	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
)

var artifactTypes = []shared.ArtifactType{
	shared.UntypedArtifact,
	shared.StringArtifact,
	shared.BoolArtifact,
	shared.NumericArtifact,
	shared.DictArtifact,
	shared.TupleArtifact,
	shared.TableArtifact,
	shared.JsonArtifact,
	shared.BytesArtifact,
	shared.ImageArtifact,
	shared.PicklableArtifact,
}

// randArtifactType generates a random artifact type.
func randArtifactType() shared.ArtifactType {
	return artifactTypes[rand.Intn(len(artifactTypes))]
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

// sampleWorkflowIDs randomly samples workflows n times with replacement,
// and returns the ID of selected Workflows.
func sampleWorkflowIDs(n int, workflows []models.Workflow) []uuid.UUID {
	workflowIDs := make([]uuid.UUID, 0, len(workflows))
	for _, workflow := range workflows {
		workflowIDs = append(workflowIDs, workflow.ID)
	}
	return sampleIDs(n, workflowIDs)
}

// sampleDagIDs randomly samples dags n times with replacement,
// and returns the ID of selected DAGs.
func sampleDagIDs(n int, dags []models.DAG) []uuid.UUID {
	dagIDs := make([]uuid.UUID, 0, len(dags))
	for _, dag := range dags {
		dagIDs = append(dagIDs, dag.ID)
	}
	return sampleIDs(n, dagIDs)
}
