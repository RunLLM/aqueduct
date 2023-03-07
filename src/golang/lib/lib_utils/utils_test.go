package lib_utils

import (
	"fmt"
	"os"
	"path/filepath"
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

func TestExtractAwsCredentials(t *testing.T) {
	credentialsFilepath := filepath.Join(t.TempDir(), "credentials_test")
	f, err := os.Create(credentialsFilepath)
	require.Nil(t, err)
	f.WriteString(
		`[default]
aws_access_key_id=dummyid
aws_secret_access_key=dummykey`,
	)
	f.Close()

	config := &shared.S3Config{
		Region:             "us-east-2",
		Bucket:             "dummybucket",
		CredentialsPath:    credentialsFilepath,
		CredentialsProfile: "default",
	}
	// Expect proper extraction
	id, key, err := ExtractAwsCredentials(config)
	require.Nil(t, err)
	require.Equal(t, "dummyid", id)
	require.Equal(t, "dummykey", key)

	config.CredentialsProfile = "user"
	// Expect error to be thrown for unknown profile
	id, key, err = ExtractAwsCredentials(config)
	require.Equal(t, "", id)
	require.Equal(t, "", key)
	require.Error(t, err)
}
