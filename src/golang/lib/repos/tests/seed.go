package tests

import (
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

const (
	// Defaults used for seeding database records
	testOrgID = "aqueduct-test"
)

// seedNotification creates count notification records for a generated user.
func (ts *TestSuite) seedNotification(count int) []models.Notification {
	notifications := make([]models.Notification, 0, count)
	users := ts.seedUser(1)
	receiverID := users[0].ID

	for i := 0; i < count; i++ {
		content := randString(10)
		level := shared.SuccessNotificationLevel
		association := &shared.NotificationAssociation{
			Object: shared.OrgNotificationObject,
			ID: uuid.New(),
		}
		notification, err := ts.notification.Create(ts.ctx, receiverID, content, level, association, ts.DB)
		require.Nil(ts.T(), err)

		notifications = append(notifications, *notification)
	}

	return notifications
}

// seedUser creates count user records.
func (ts *TestSuite) seedUser(count int) []models.User {
	users := make([]models.User, 0, count)

	for i := 0; i < count; i++ {
		user, err := ts.user.Create(ts.ctx, testOrgID, randAPIKey(), ts.DB)
		require.Nil(ts.T(), err)

		users = append(users, *user)
	}

	return users
}

// seedWorkflow creates count workflow records.
// It creates a new user as the workflows' owner.
func (ts *TestSuite) seedWorkflow(count int) []models.Workflow {
	users := ts.seedUser(1)
	userIDs := sampleUserIDs(count, users)
	return ts.seedWorkflowWithUser(count, userIDs)
}

// seedWorkflowWithUser creates count workflow records. It uses userIDs as the
// owner of each workflow.
func (ts *TestSuite) seedWorkflowWithUser(count int, userIDs []uuid.UUID) []models.Workflow {
	require.Len(ts.T(), userIDs, count)

	workflows := make([]models.Workflow, 0, count)

	for i := 0; i < count; i++ {
		userID := userIDs[i]
		name := randString(10)
		description := randString(15)
		schedule := &shared.Schedule{
			Trigger:              shared.PeriodicUpdateTrigger,
			CronSchedule:         "* * * * *",
			DisableManualTrigger: false,
			Paused:               false,
		}
		retentionPolicy := &shared.RetentionPolicy{
			KLatestRuns: 10,
		}

		workflow, err := ts.workflow.Create(
			ts.ctx,
			userID,
			name,
			description,
			schedule,
			retentionPolicy,
			ts.DB,
		)
		require.Nil(ts.T(), err)

		workflows = append(workflows, *workflow)
	}

	return workflows
}
