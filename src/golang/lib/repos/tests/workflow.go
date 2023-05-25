package tests

import (
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
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

func (ts *TestSuite) TestWorkflow_GetByScheduleTrigger() {
	triggerWorkflow := ts.seedWorkflow(1)[0]

	expectedWorkflows := make([]models.Workflow, 0, 2)
	for i := 0; i < 2; i++ {
		workflow, err := ts.workflow.Create(
			ts.ctx,
			triggerWorkflow.UserID,
			randString(10),
			randString(15),
			&shared.Schedule{
				Trigger:  shared.CascadingUpdateTrigger,
				SourceID: triggerWorkflow.ID,
			},
			&shared.RetentionPolicy{
				KLatestRuns: 5,
			},
			&shared.NotificationSettings{},
			ts.DB,
		)
		require.Nil(ts.T(), err)

		expectedWorkflows = append(expectedWorkflows, *workflow)
	}

	actualWorkflows, err := ts.workflow.GetByScheduleTrigger(ts.ctx, shared.CascadingUpdateTrigger, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqualWorkflows(ts.T(), expectedWorkflows, actualWorkflows)
}

func (ts *TestSuite) TestWorkflow_GetTargets() {
	triggerWorkflow := ts.seedWorkflow(1)[0]

	// Create 2 Workflows where the schedule is a cascading update after
	// `triggerWorkflow` completes
	expectedIDs := make([]uuid.UUID, 0, 2)
	for i := 0; i < 2; i++ {
		workflow, err := ts.workflow.Create(
			ts.ctx,
			triggerWorkflow.UserID,
			randString(10),
			randString(15),
			&shared.Schedule{
				Trigger:  shared.CascadingUpdateTrigger,
				SourceID: triggerWorkflow.ID,
			},
			&shared.RetentionPolicy{
				KLatestRuns: 5,
			},
			&shared.NotificationSettings{},
			ts.DB,
		)
		require.Nil(ts.T(), err)

		expectedIDs = append(expectedIDs, workflow.ID)
	}

	actualIDs, err := ts.workflow.GetTargets(ts.ctx, triggerWorkflow.ID, ts.DB)
	require.Nil(ts.T(), err)
	require.ElementsMatch(ts.T(), expectedIDs, actualIDs)
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
	notificationResourceID := uuid.New()

	expectedWorkflow := &models.Workflow{
		UserID:      user.ID,
		Name:        randString(10),
		Description: randString(20),
		Schedule: shared.Schedule{
			Trigger:              shared.PeriodicUpdateTrigger,
			CronSchedule:         "* * * * *",
			DisableManualTrigger: false,
			Paused:               false,
		},
		RetentionPolicy: shared.RetentionPolicy{
			KLatestRuns: 10,
		},
		NotificationSettings: shared.NotificationSettings{
			Settings: map[uuid.UUID]shared.NotificationLevel{
				notificationResourceID: shared.ErrorNotificationLevel,
			},
		},
	}

	actualWorkflow, err := ts.workflow.Create(
		ts.ctx,
		expectedWorkflow.UserID,
		expectedWorkflow.Name,
		expectedWorkflow.Description,
		&expectedWorkflow.Schedule,
		&expectedWorkflow.RetentionPolicy,
		&expectedWorkflow.NotificationSettings,
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
	notificationResourceID := uuid.New()

	newName := "new_workflow_name"
	newSchedule := shared.Schedule{
		Trigger:              shared.ManualUpdateTrigger,
		CronSchedule:         "0 0 0 0 0",
		DisableManualTrigger: true,
		Paused:               true,
	}

	newNotificationSettings := shared.NotificationSettings{
		Settings: map[uuid.UUID]shared.NotificationLevel{
			notificationResourceID: shared.ErrorNotificationLevel,
		},
	}

	changes := map[string]interface{}{
		models.WorkflowName:                 newName,
		models.WorkflowSchedule:             &newSchedule,
		models.WorkflowNotificationSettings: &newNotificationSettings,
	}

	newWorkflow, err := ts.workflow.Update(ts.ctx, oldWorkflow.ID, changes, ts.DB)
	require.Nil(ts.T(), err)

	requireDeepEqual(ts.T(), newSchedule, newWorkflow.Schedule)
	require.Equal(ts.T(), newName, newWorkflow.Name)
	requireDeepEqual(ts.T(), newWorkflow.NotificationSettings, newNotificationSettings)
}

func (ts *TestSuite) TestWorkflow_RemoveNotificationFromSettings() {
	users := ts.seedUser(1)
	user := users[0]
	notificationToRemove := uuid.New()
	notificationToKeep := uuid.New()

	workflow := &models.Workflow{
		UserID: user.ID,
		NotificationSettings: shared.NotificationSettings{
			Settings: map[uuid.UUID]shared.NotificationLevel{
				notificationToRemove: shared.ErrorNotificationLevel,
				notificationToKeep:   shared.SuccessNotificationLevel,
			},
		},
	}

	workflowCreated, err := ts.workflow.Create(
		ts.ctx,
		workflow.UserID,
		workflow.Name,
		workflow.Description,
		&workflow.Schedule,
		&workflow.RetentionPolicy,
		&workflow.NotificationSettings,
		ts.DB,
	)
	require.Nil(ts.T(), err)

	err = ts.workflow.RemoveNotificationFromSettings(ts.ctx, notificationToRemove, ts.DB)
	require.Nil(ts.T(), err)
	expectedSettings := shared.NotificationSettings{
		Settings: map[uuid.UUID]shared.NotificationLevel{
			notificationToKeep: shared.SuccessNotificationLevel,
		},
	}

	actualWorkflow, err := ts.workflow.Get(ts.ctx, workflowCreated.ID, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqual(ts.T(), expectedSettings, actualWorkflow.NotificationSettings)
}
