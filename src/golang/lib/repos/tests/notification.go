package tests

import (
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func (ts *TestSuite) TestNotification_GetByReceiverAndStatus() {
	expectedNotifications := ts.seedNotification(1)
	expectedNotification := &expectedNotifications[0]

	actualNotification, err := ts.notification.GetByReceiverAndStatus(ts.ctx, expectedNotification.ReceiverID, shared.UnreadNotificationStatus, ts.DB)
	require.Nil(ts.T(), err)
	requireDeepEqual(ts.T(), expectedNotifications, actualNotification)
}

func (ts *TestSuite) TestNotification_ValidateUser() {
	notifications := ts.seedNotification(1)
	notification := &notifications[0]

	validateTrue, validateTrueErr := ts.notification.ValidateUser(ts.ctx, notification.ID, notification.ReceiverID, ts.DB)
	require.Nil(ts.T(), validateTrueErr)
	require.True(ts.T(), validateTrue)

	validateFalse, validateFalseErr := ts.notification.ValidateUser(ts.ctx, notification.ID, uuid.New(), ts.DB)
	require.Nil(ts.T(), validateFalseErr)
	require.False(ts.T(), validateFalse)
}

func (ts *TestSuite) TestNotification_Create() {
	users := ts.seedUser(1)
	receiverID := users[0].ID
	content := randString(10)
	level := shared.SuccessNotificationLevel
	association := &shared.NotificationAssociation{
		Object: shared.OrgNotificationObject,
		ID:     uuid.New(),
	}
	expectedNotification := &models.Notification{
		ReceiverID:  receiverID,
		Content:     content,
		Status:      shared.UnreadNotificationStatus,
		Level:       level,
		Association: *association,
	}

	actualNotification, err := ts.notification.Create(ts.ctx, receiverID, content, level, association, ts.DB)
	require.Nil(ts.T(), err)

	require.NotEqual(ts.T(), uuid.Nil, actualNotification.ID)

	expectedNotification.ID = actualNotification.ID
	expectedNotification.CreatedAt = actualNotification.CreatedAt
	requireDeepEqual(ts.T(), expectedNotification, actualNotification)
}

func (ts *TestSuite) TestNotification_Update() {
	notifications := ts.seedNotification(1)
	notification := &notifications[0]

	unreadNotification, unreadErr := ts.notification.Update(ts.ctx, notification.ID, shared.UnreadNotificationStatus, ts.DB)
	require.Nil(ts.T(), unreadErr)
	require.Equal(ts.T(), unreadNotification.Status, shared.UnreadNotificationStatus)

	archivedNotification, archivedErr := ts.notification.Update(ts.ctx, notification.ID, shared.ArchivedNotificationStatus, ts.DB)
	require.Nil(ts.T(), archivedErr)
	require.Equal(ts.T(), archivedNotification.Status, shared.ArchivedNotificationStatus)
}
