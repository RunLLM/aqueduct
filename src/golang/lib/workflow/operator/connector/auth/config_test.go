package auth

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseConfigType(t *testing.T) {
	testCases := []struct {
		description    string
		input          string
		expectedOutput configType
	}{
		{
			description:    "staticConfig",
			input:          "staticConfig",
			expectedOutput: staticConfigType,
		},
		{
			description:    "oauthConfig",
			input:          "oauthConfig",
			expectedOutput: oauthConfigType,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			output, err := parseConfigType(tc.input)
			require.Nil(t, err)
			require.Equal(t, tc.expectedOutput, output)
		})
	}
}
