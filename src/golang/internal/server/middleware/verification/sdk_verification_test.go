package verification_test

import (
	"net/http"
	"testing"

	"github.com/aqueducthq/aqueduct/internal/server/middleware/verification"
	"github.com/stretchr/testify/require"
)

func TestVerifySdkRequest(t *testing.T) {
	responseCode, reason := verification.VerifySdkRequest("100000000")
	require.Equal(t, reason, "Sdk client version accepted")
	require.Equal(t, responseCode, http.StatusOK)

	responseCode, reason = verification.VerifySdkRequest("0")
	require.Equal(t, reason, "Sdk client is not supported. Please upgrade to supported versions.")
	require.Equal(t, responseCode, http.StatusForbidden)

	responseCode, reason = verification.VerifySdkRequest("log(0)")
	require.Equal(t, reason, "Could not recognize the recieved sdk client version as an integer")
	require.Equal(t, responseCode, http.StatusBadRequest)
}
