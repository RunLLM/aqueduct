package tests

import (
	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func (ts *TestSuite) TestWorkflow_Exists() {
	workflows := ts.seedWorkflow(1)
	workflow := workflows[0]

	exists, err := ts.workflow.Exists(ts.ctx, workflow.ID, ts.DB)
	require.Nil(ts.T(), err)
	require.True(ts.T(), exists)

	// Check for non-existent workflow
	exists, err = ts.workflow.Exists(ts.ctx, uuid.Nil, ts.DB)
	require.Nil(ts.T(), err)
	require.False(ts.T(), exists)
}

func (ts *TestSuite) TestWorkflow_Get() {
	workflows := ts.seedWorkflow(1)
	expectedWorkflow := workflows[0]

	actualWorkflow, err := ts.workflow.Get(ts.ctx, expectedWorkflow.ID, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqual(ts.T(), expectedWorkflow, *actualWorkflow)
}

func (ts *TestSuite) TestWorkflow_GetByOwnerAndName() {
	users := ts.seedUser(1)
	user := users[0]

	workflows := ts.seedWorkflowWithUser(1, []uuid.UUID{user.ID})
	workflow := &workflows[0]

	actualWorkflow, err := ts.workflow.GetByOwnerAndName(ts.ctx, user.ID, workflow.Name, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqual(ts.T(), workflow, actualWorkflow)
}

func (ts *TestSuite) TestWorkflow_GetLatestStatusesByOrg() {
	// TODO: Implement once DAG, DAGRun, and DAGResult collections are refactored
}

func (ts *TestSuite) TestWorkflow_List() {
	workflows := ts.seedWorkflow(2)

	actualWorkflows, err := ts.workflow.List(ts.ctx, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqualWorkflows(ts.T(), workflows, actualWorkflows)
}

func (ts *TestSuite) TestWorkflow_ValidateOrg() {
	users := ts.seedUser(1)
	user := users[0]

	workflows := ts.seedWorkflowWithUser(1, []uuid.UUID{user.ID})
	workflow := &workflows[0]

	validOrg, err := ts.workflow.ValidateOrg(ts.ctx, workflow.ID, user.OrgID, ts.DB)
	require.Nil(ts.T(), err)
	require.True(ts.T(), validOrg)
}

func (ts *TestSuite) TestWorkflow_Create() {
	users := ts.seedUser(1)
	user := users[0]

	expectedWorkflow := &models.Workflow{
		UserID:      user.ID,
		Name:        randString(10),
		Description: randString(20),
		Schedule: workflow.Schedule{
			Trigger:              workflow.PeriodicUpdateTrigger,
			CronSchedule:         "* * * * *",
			DisableManualTrigger: false,
			Paused:               false,
		},
		RetentionPolicy: workflow.RetentionPolicy{
			KLatestRuns: 10,
		},
	}

	actualWorkflow, err := ts.workflow.Create(
		ts.ctx,
		expectedWorkflow.UserID,
		expectedWorkflow.Name,
		expectedWorkflow.Description,
		&expectedWorkflow.Schedule,
		&expectedWorkflow.RetentionPolicy,
		ts.DB,
	)
	require.Nil(ts.T(), err)

	require.NotEqual(ts.T(), uuid.Nil, actualWorkflow.ID)

	expectedWorkflow.ID = actualWorkflow.ID
	expectedWorkflow.CreatedAt = actualWorkflow.CreatedAt
	requireDeepEqual(ts.T(), expectedWorkflow, actualWorkflow)
}

func (ts *TestSuite) TestWorkflow_Delete() {
	workflows := ts.seedWorkflow(1)
	workflow := workflows[0]

	err := ts.workflow.Delete(ts.ctx, workflow.ID, ts.DB)
	require.Nil(ts.T(), err)
}

func (ts *TestSuite) TestWorkflow_Update() {
	workflows := ts.seedWorkflow(1)
	oldWorkflow := workflows[0]

	newName := "new_workflow_name"
	newSchedule := workflow.Schedule{
		Trigger:              workflow.ManualUpdateTrigger,
		CronSchedule:         "0 0 0 0 0",
		DisableManualTrigger: true,
		Paused:               true,
	}

	changes := map[string]interface{}{
		models.WorkflowName:     newName,
		models.WorkflowSchedule: &newSchedule,
	}

	newWorkflow, err := ts.workflow.Update(ts.ctx, oldWorkflow.ID, changes, ts.DB)
	require.Nil(ts.T(), err)

	requireDeepEqual(ts.T(), newSchedule, newWorkflow.Schedule)
	require.Equal(ts.T(), newName, newWorkflow.Name)
}
