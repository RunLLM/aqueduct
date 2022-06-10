package tests

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"

	"github.com/aqueducthq/aqueduct/lib/collections/user"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func requireDeepEqual(t *testing.T, expected, actual interface{}) {
	require.True(
		t,
		reflect.DeepEqual(
			expected,
			actual,
		),
		fmt.Sprintf("Expected: %v\n Actual: %v", expected, actual),
	)
}

// randString returns a random string of length n.
func randString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))] // nolint:gosec
	}
	return string(b)
}

// randIdsFromList randomly polls ids n times.
func randIdsFromList(n int, ids []uuid.UUID) []uuid.UUID {
	randIds := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		randId := ids[rand.Intn(len(ids))] // nolint:gosec
		randIds = append(randIds, randId)
	}
	return randIds
}

// randUserIdsFromList randomly polls users n times and returns
// the userIds selected.
func randUserIdsFromList(n int, users []user.User) []uuid.UUID {
	userIds := make([]uuid.UUID, 0, len(users))
	for _, userObj := range users {
		userIds = append(userIds, userObj.Id)
	}
	return randIdsFromList(n, userIds)
}

// randWorkflowIdsFromList randomly polls workflows n times and returns
// the workflowIds selected.
func randWorkflowIdsFromList(n int, workflows []workflow.Workflow) []uuid.UUID {
	workflowIds := make([]uuid.UUID, 0, len(workflows))
	for _, workflowObj := range workflows {
		workflowIds = append(workflowIds, workflowObj.Id)
	}
	return randIdsFromList(n, workflowIds)
}

// randWorkflowDagIdsFromList randomly polls workflowDags n times and returns
// the workflowDagIds selected.
func randWorkflowDagIdsFromList(n int, workflowDags []workflow_dag.WorkflowDag) []uuid.UUID {
	workflowDagIds := make([]uuid.UUID, 0, len(workflowDags))
	for _, workflowDagObj := range workflowDags {
		workflowDagIds = append(workflowDagIds, workflowDagObj.Id)
	}
	return randIdsFromList(n, workflowDagIds)
}

// randOrgIdsFromList randomly polls users n times and returns
// the orgIds selected.
func randOrgIdsFromList(n int, users []user.User) []string {
	orgIds := make([]string, 0, n)
	for i := 0; i < n; i++ {
		idx := rand.Intn(len(users)) // nolint:gosec
		orgIds = append(orgIds, users[idx].OrganizationId)
	}
	return orgIds
}
