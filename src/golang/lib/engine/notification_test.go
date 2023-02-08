// Note about package and file structure for testings:
// https://medium.com/@butterv/go-how-to-implement-tests-of-private-methods-e34d1cc2bc31
package engine

import (
	"testing"

	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/stretchr/testify/require"
)

func TestGetNotification(t *testing.T) {
	expectedMsg := "Workflow example_dag succeeded."
	actualMsg := notificationMsg("example_dag", shared.SuccessNotificationLevel, "")
	require.Equal(t, expectedMsg, actualMsg)

	expectedMsg = "Workflow example_dag succeeded: example_ctx ."
	actualMsg = notificationMsg("example_dag", shared.SuccessNotificationLevel, "example_ctx")
	require.Equal(t, expectedMsg, actualMsg)

	expectedMsg = "Workflow example_dag succeeded with warning."
	actualMsg = notificationMsg("example_dag", shared.WarningNotificationLevel, "")
	require.Equal(t, expectedMsg, actualMsg)

	expectedMsg = "Workflow example_dag succeeded with warning: example_ctx ."
	actualMsg = notificationMsg("example_dag", shared.WarningNotificationLevel, "example_ctx")
	require.Equal(t, expectedMsg, actualMsg)

	expectedMsg = "Workflow example_dag failed."
	actualMsg = notificationMsg("example_dag", shared.ErrorNotificationLevel, "")
	require.Equal(t, expectedMsg, actualMsg)

	expectedMsg = "Workflow example_dag failed: example_ctx ."
	actualMsg = notificationMsg("example_dag", shared.ErrorNotificationLevel, "example_ctx")
	require.Equal(t, expectedMsg, actualMsg)

	expectedMsg = "Workflow example_dag has a message: example_ctx ."
	actualMsg = notificationMsg("example_dag", shared.NeutralNotificationLevel, "example_ctx")
	require.Equal(t, expectedMsg, actualMsg)

	expectedMsg = "Workflow example_dag has a message: example_ctx ."
	actualMsg = notificationMsg("example_dag", shared.InfoNotificationLevel, "example_ctx")
	require.Equal(t, expectedMsg, actualMsg)
}
