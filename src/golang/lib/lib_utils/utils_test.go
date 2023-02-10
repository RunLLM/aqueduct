package lib_utils

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/workflow/operator/connector/auth"
	"github.com/stretchr/testify/require"
)

// TODO: This should be merged with `requireDeepEqual` in `repo` package.
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

func TestParseEmailConfig(t *testing.T) {
	configMap := map[string]string{
		"user":               "test_user",
		"password":           "test_password",
		"host":               "test_host",
		"port":               "test_port",
		"targets_serialized": "[\"test_target_1\", \"test_target_2\"]",
		"level":              "warning",
		"enabled":            "false",
	}

	staticConfig := auth.NewStaticConfig(configMap)

	expectedConfig := &shared.EmailConfig{
		User:     configMap["user"],
		Password: configMap["password"],
		Host:     configMap["host"],
		Port:     configMap["port"],
		Targets:  []string{"test_target_1", "test_target_2"},
		Level:    shared.WarningNotificationLevel,
		Enabled:  false,
	}

	actualConfig, err := ParseEmailConfig(staticConfig)
	require.Nil(t, err)
	requireDeepEqual(t, expectedConfig, actualConfig)
}

func TestParseSlackConfig(t *testing.T) {
	configMap := map[string]string{
		"token":               "test_token",
		"channels_serialized": "[\"channel_1\", \"channel_2\"]",
		"level":               "success",
		"enabled":             "true",
	}

	staticConfig := auth.NewStaticConfig(configMap)

	expectedConfig := &shared.SlackConfig{
		Token:    configMap["token"],
		Channels: []string{"channel_1", "channel_2"},
		Level:    shared.SuccessNotificationLevel,
		Enabled:  true,
	}

	actualConfig, err := ParseSlackConfig(staticConfig)
	require.Nil(t, err)
	requireDeepEqual(t, expectedConfig, actualConfig)
}
