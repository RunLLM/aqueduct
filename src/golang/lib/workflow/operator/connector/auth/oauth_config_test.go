package auth

import (
	"testing"

	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func TestOAuthConfig(t *testing.T) {
	testCases := []struct {
		description        string
		accessToken        string
		publicConf         map[string]string
		expectedMarshalStr string
	}{
		{
			description: "basic",
			accessToken: "12345",
			publicConf: map[string]string{
				"email": "test@aqueducthq.com",
			},
			expectedMarshalStr: `{"access_token":"12345","email":"test@aqueducthq.com"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			config := &OAuthConfig{
				Token:      &oauth2.Token{AccessToken: tc.accessToken},
				PublicConf: tc.publicConf,
			}
			testOAuthConfigGetType(t, config)
			testOAuthConfigPublicConfig(t, config, tc.publicConf)
			testOAuthConfigMarshal(t, config, tc.expectedMarshalStr)
		})
	}
}

func testOAuthConfigGetType(t *testing.T, config *OAuthConfig) {
	require.Equal(t, oauthConfigType, config.getType())
}

func testOAuthConfigPublicConfig(t *testing.T, config *OAuthConfig, expectedPublicConfig map[string]string) {
	require.Equal(t, expectedPublicConfig, config.PublicConfig())
}

func testOAuthConfigMarshal(t *testing.T, config *OAuthConfig, expectedMarshalStr string) {
	data, err := config.Marshal()
	require.Nil(t, err)
	require.Equal(t, expectedMarshalStr, string(data))
}

func TestNewOAuth2Config(t *testing.T) {
	testCases := []struct {
		description          string
		service              shared.Service
		clientId             string
		clientSecret         string
		redirectURL          string
		expectedOAuth2Config *oauth2.Config
		expectsErr           bool
	}{
		{
			description:  "google sheets",
			service:      shared.GoogleSheets,
			clientId:     "testClient",
			clientSecret: "secret",
			redirectURL:  "postmessage",
			expectedOAuth2Config: &oauth2.Config{
				ClientID:     "testClient",
				ClientSecret: "secret",
				Endpoint:     google.Endpoint,
				RedirectURL:  "postmessage",
				Scopes:       []string{googleContactsReadOnlyScope, googleSheetsScope, googleDriveScope},
			},
			expectsErr: false,
		},
		{
			description:          "non-OAuth2 supported service",
			service:              shared.Postgres,
			clientId:             "testClient",
			clientSecret:         "secret",
			expectedOAuth2Config: nil,
			expectsErr:           true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			oauthConf, err := newOAuth2Config(tc.service, tc.clientId, tc.clientSecret, tc.redirectURL)
			require.Equal(t, tc.expectedOAuth2Config, oauthConf)
			if tc.expectsErr {
				require.NotNil(t, err)
			} else {
				require.Nil(t, err)
			}
		})
	}
}
