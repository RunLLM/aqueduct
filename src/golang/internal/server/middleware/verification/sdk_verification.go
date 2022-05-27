package verification

import (
	"net/http"
	"strconv"

	"github.com/aqueducthq/aqueduct/internal/server/utils"
)

// Verifies whether a request coming from the SDK client is valid
// Current requirements are:
// 1) Ensure client version is greater or equal to the allowed version: utils.Constants.AllowedSdkClientVersion
func VerifySdkRequest(sdkVersion string) (responseCode int, reason string) {
	sdkVersionParsed, err := strconv.Atoi(sdkVersion)
	if err != nil {
		return http.StatusBadRequest, "Could not recognize the recieved sdk client version as an integer"
	}

	if sdkVersionParsed < utils.AllowedSdkClientVersion {
		return http.StatusForbidden, "Sdk client is not supported. Please upgrade to supported versions."
	}

	return http.StatusOK, "Sdk client version accepted"
}
