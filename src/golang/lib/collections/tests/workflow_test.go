package tests

import (
	"context"
	"testing"
	"time"

	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func seedWorkflow(t *testing.T, count int) []workflow.Workflow {
	numUsers := 2

	users := seedUser(t, numUsers)
	userIds := randUserIdsFromList(count, users)

	return seedWorkflowWithUsers(t, count, userIds)
}

// seedWorkflowWithUsers populates the workflow table with count workflows where
// user_id is set to the values provided in userIds.
func seedWorkflowWithUsers(t *testing.T, count int, userIds []uuid.UUID) []workflow.Workflow {
	require.Len(t, userIds, count)

	workflows := make([]workflow.Workflow, 0, count)

	for i := 0; i < count; i++ {
		userId := userIds[i]
		name := randString(10)
		description := randString(15)
		schedule := workflow.Schedule{
			Trigger:              workflow.ManualUpdateTrigger,
			CronSchedule:         "* * * * *",
			DisableManualTrigger: false,
		}
		retentionPolicy := workflow.RetentionPolicy{
			KLatestRuns: 10,
		}

		testWorkflow, err := writers.workflowWriter.CreateWorkflow(
			context.Background(),
			userId,
			name,
			description,
			&schedule,
			&retentionPolicy,
			db,
		)
		require.Nil(t, err)

		workflows = append(workflows, *testWorkflow)
	}

	require.Len(t, workflows, count)

	return workflows
}

func requireEqualWorkflows(t *testing.T, expected, actual []workflow.Workflow) {
	require.Equal(t, len(expected), len(actual))

	for _, expectedWorkflow := range expected {
		found := false
		for _, actualWorkflow := range actual {
			if expectedWorkflow.Id == actualWorkflow.Id {
				found = true
				requireDeepEqual(t, expectedWorkflow, actualWorkflow)
			}
		}
		require.True(t, found, "Unable to find workflow: %v", expectedWorkflow)
	}
}

// idsFromWorkflows returns the ids from the workflows provided.
func idsFromWorkflows(workflows []workflow.Workflow) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(workflows))
	for _, workflowObj := range workflows {
		ids = append(ids, workflowObj.Id)
	}
	return ids
}

func TestCreateWorkflow(t *testing.T) {
	defer resetDatabase(t)

	users := seedUser(t, 1)

	expectedWorkflow := &workflow.Workflow{
		UserId:      users[0].Id,
		Name:        "test",
		Description: "testing flow",
		Schedule: workflow.Schedule{
			Trigger:              workflow.ManualUpdateTrigger,
			CronSchedule:         "* * * * *",
			DisableManualTrigger: false,
		},
		RetentionPolicy: workflow.RetentionPolicy{KLatestRuns: 10},
	}
	createdAtFloor := time.Now()

	actualWorkflow, err := writers.workflowWriter.CreateWorkflow(
		context.Background(),
		expectedWorkflow.UserId,
		expectedWorkflow.Name,
		expectedWorkflow.Description,
		&expectedWorkflow.Schedule,
		&expectedWorkflow.RetentionPolicy,
		db,
	)
	require.Nil(t, err)
	require.NotEqual(t, uuid.Nil, actualWorkflow.Id)
	require.True(t, actualWorkflow.CreatedAt.After(createdAtFloor))

	expectedWorkflow.Id = actualWorkflow.Id
	expectedWorkflow.CreatedAt = actualWorkflow.CreatedAt

	requireDeepEqual(t, expectedWorkflow, actualWorkflow)
}

func TestWorkflowExists(t *testing.T) {
	defer resetDatabase(t)

	numWorkflows := 1

	workflows := seedWorkflow(t, numWorkflows)
	require.Len(t, workflows, numWorkflows)

	exists, err := readers.workflowReader.Exists(context.Background(), workflows[0].Id, db)
	require.Nil(t, err)
	require.True(t, exists)

	exists, err = readers.workflowReader.Exists(context.Background(), uuid.Nil, db)
	require.Nil(t, err)
	require.False(t, exists)
}

func TestGetWorkflow(t *testing.T) {
	defer resetDatabase(t)

	numWorkflow := 3

	workflows := seedWorkflow(t, numWorkflow)

	testWorkflows := workflows[:2]
	testWorkflowIds := idsFromWorkflows(testWorkflows)

	actualWorkflows, err := readers.workflowReader.GetWorkflows(context.Background(), testWorkflowIds, db)
	require.Nil(t, err)
	requireEqualWorkflows(t, testWorkflows, actualWorkflows)
}

func TestGetWorkflowsByUser(t *testing.T) {
	defer resetDatabase(t)

	numUsers := 1
	numOtherUsers := 2
	numWorkflows := 3
	numOtherWorkflows := 2

	testUsers := seedUser(t, numUsers)
	testUserIds := randUserIdsFromList(numWorkflows, testUsers)

	otherUsers := seedUser(t, numOtherUsers)
	otherUserIds := randUserIdsFromList(numOtherWorkflows, otherUsers)

	// Create workflows by test user
	testWorkflows := seedWorkflowWithUsers(t, numWorkflows, testUserIds)

	// Create workflows by "other" users
	seedWorkflowWithUsers(t, numOtherWorkflows, otherUserIds)

	workflows, err := readers.workflowReader.GetWorkflowsByUser(context.Background(), testUsers[0].Id, db)
	require.Nil(t, err)
	requireEqualWorkflows(t, testWorkflows, workflows)
}

func TestUpdateWorkflow(t *testing.T) {
	defer resetDatabase(t)

	numWorkflows := 1

	testWorkflows := seedWorkflow(t, numWorkflows)
	testWorkflow := testWorkflows[0]

	testWorkflow.Description = "this is a new description"
	testWorkflow.Schedule = workflow.Schedule{
		Trigger:              workflow.PeriodicUpdateTrigger,
		CronSchedule:         "1 * * * *",
		DisableManualTrigger: true,
	}

	newWorkflow, err := writers.workflowWriter.UpdateWorkflow(
		context.Background(),
		testWorkflow.Id,
		map[string]interface{}{
			workflow.DescriptionColumn: testWorkflow.Description,
			workflow.ScheduleColumn:    &testWorkflow.Schedule,
		},
		db,
	)
	require.Nil(t, err)
	requireDeepEqual(t, testWorkflow, *newWorkflow)
}
