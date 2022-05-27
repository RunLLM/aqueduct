package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStaticConfig(t *testing.T) {
	testCases := []struct {
		description          string
		configMap            map[string]string
		expectedPublicConfig map[string]string
		expectedMarshalStr   string
	}{
		{
			description: "basic",
			configMap: map[string]string{
				"username":                    "test",
				"password":                    "password",
				"service_account_credentials": "cred",
				"database":                    "test-db",
			},
			expectedPublicConfig: map[string]string{
				"username": "test",
				"database": "test-db",
			},
			expectedMarshalStr: `{"database":"test-db","password":"password","service_account_credentials":"cred","username":"test"}`,
		},
		{
			description: "no public config",
			configMap: map[string]string{
				"password":                    "password",
				"service_account_credentials": "cred",
			},
			expectedPublicConfig: map[string]string{},
			expectedMarshalStr:   `{"password":"password","service_account_credentials":"cred"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			config := NewStaticConfig(tc.configMap)
			testStaticConfigGetType(t, config)
			testStaticConfigPublicConfig(t, config, tc.expectedPublicConfig)
			testStaticConfigMarshal(t, config, tc.expectedMarshalStr)
			testStaticConfigRefresh(t, config)
		})
	}
}

func testStaticConfigGetType(t *testing.T, config Config) {
	require.Equal(t, staticConfigType, config.getType())
}

func testStaticConfigPublicConfig(t *testing.T, config Config, expectedPublicConfig map[string]string) {
	publicConfig := config.PublicConfig()
	require.Equal(t, expectedPublicConfig, publicConfig)
}

func testStaticConfigMarshal(t *testing.T, config Config, expectedMarshalStr string) {
	data, err := config.Marshal()
	require.Nil(t, err)
	require.Equal(t, expectedMarshalStr, string(data))
}

func testStaticConfigRefresh(t *testing.T, config Config) {
	refresh, err := config.Refresh(context.Background())
	require.Nil(t, err)
	require.False(t, refresh)
}
