package notification

import (
	"testing"

	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/stretchr/testify/require"
)

func TestShouldSend(t *testing.T) {
	severities := []shared.NotificationLevel{
		shared.SuccessNotificationLevel,
		shared.WarningNotificationLevel,
		shared.ErrorNotificationLevel,
		shared.InfoNotificationLevel,
		shared.NeutralNotificationLevel,
	}

	for _, severity := range severities {
		require.Equal(t, true, ShouldSend(severity, severity))

		// success, info and neutral thresholds always let through regardless of level
		require.Equal(t, true, ShouldSend(shared.InfoNotificationLevel, severity))
		require.Equal(t, true, ShouldSend(shared.NeutralNotificationLevel, severity))
		require.Equal(t, true, ShouldSend(shared.SuccessNotificationLevel, severity))

		// info and neutral always get through regardless of threshold
		require.Equal(t, true, ShouldSend(severity, shared.InfoNotificationLevel))
		require.Equal(t, true, ShouldSend(severity, shared.NeutralNotificationLevel))
	}

	require.Equal(t, false, ShouldSend(shared.ErrorNotificationLevel, shared.WarningNotificationLevel))
	require.Equal(t, false, ShouldSend(shared.ErrorNotificationLevel, shared.SuccessNotificationLevel))
	require.Equal(t, true, ShouldSend(shared.WarningNotificationLevel, shared.ErrorNotificationLevel))
	require.Equal(t, false, ShouldSend(shared.WarningNotificationLevel, shared.SuccessNotificationLevel))
}
