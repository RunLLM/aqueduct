package tests

import (
	"context"
	"testing"
	"time"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/collections/utils"
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

func TestGetWorkflowsWithLatestRunResult(t *testing.T) {
	defer resetDatabase(t)

	// Create 2 test workflows
	testWorkflows := seedWorkflow(t, 2)
	testWorkflow1, testWorkflow2 := testWorkflows[0], testWorkflows[1]

	// Create DAG for workflow 1
	testWorkflow1DAGs := seedWorkflowDagWithWorkflows(t, 1, []uuid.UUID{testWorkflow1.Id})
	testWorkflow1DAG := testWorkflow1DAGs[0]

	// Create DAG for workflow 2
	seedWorkflowDagWithWorkflows(t, 1, []uuid.UUID{testWorkflow2.Id})

	// Create DAG result for workflow 1 only
	testWorkflow1Results := seedWorkflowDagResultWithDags(t, 1, []uuid.UUID{testWorkflow1DAG.Id})
	testWorkflow1Result := testWorkflow1Results[0]

	latestResults, err := readers.workflowReader.GetWorkflowsWithLatestRunResult(context.Background(), testOrganizationId, db)
	require.Nil(t, err)
	require.Len(t, latestResults, 2)

	expectedResults := []workflow.LatestWorkflowResponse{
		{
			Id:          testWorkflow1.Id,
			Name:        testWorkflow1.Name,
			Description: testWorkflow1.Description,
			CreatedAt:   testWorkflow1.CreatedAt,
			LastRunAt: utils.NullTime{
				Time:   testWorkflow1Result.CreatedAt,
				IsNull: false,
			},
			Status: shared.NullExecutionStatus{
				ExecutionStatus: testWorkflow1Result.Status,
				IsNull:          false,
			},
			Engine: testWorkflow1.Engine,
		},
		{
			Id:          testWorkflow2.Id,
			Name:        testWorkflow2.Name,
			Description: testWorkflow2.Description,
			CreatedAt:   testWorkflow2.CreatedAt,
			LastRunAt: utils.NullTime{
				IsNull: true,
			},
			Status: shared.NullExecutionStatus{
				IsNull: true,
			},
			Engine: testWorkflow2.Engine,
		},
	}

	for _, expectedResult := range expectedResults {
		foundMatch := false

		for _, actualResult := range latestResults {
			if expectedResult.Id == actualResult.Id {
				requireDeepEqual(t, expectedResult, actualResult)
				foundMatch = true
				break
			}
		}

		if !foundMatch {
			t.Errorf("Unable to find matching result for workflow %v", expectedResult.Id)
			t.FailNow()
		}
	}
}
