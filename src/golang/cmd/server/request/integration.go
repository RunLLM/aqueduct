package request

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	"github.com/dropbox/godropbox/errors"
)

// ParseIntegrationConfigFromRequest parses the integration service, integration name, configuration,
// and whether it is a user only integration from the request
func ParseIntegrationConfigFromRequest(r *http.Request) (integration.Service, string, map[string]string, bool, error) {
	serviceStr := r.Header.Get(routes.IntegrationServiceHeader)
	service, err := integration.ParseService(serviceStr)
	if err != nil {
		return "", "", nil, false, err
	}

	configHeader := r.Header.Get(routes.IntegrationConfigHeader)
	var configuration map[string]string
	err = json.Unmarshal([]byte(configHeader), &configuration)
	if err != nil {
		return "", "", nil, false, errors.Newf("Unable to parse integration configuration: %v", err)
	}

	integrationName := r.Header.Get(routes.IntegrationNameHeader)
	if integrationName == "" {
		return "", "", nil, false, errors.New("Integration name was not provided.")
	}

	userOnly := isUserOnlyIntegration(service)

	return service, integrationName, configuration, userOnly, nil
}

// isUserOnlyIntegration returns whether the specified service is only accessible by the user.
func isUserOnlyIntegration(svc integration.Service) bool {
	userSpecific := []integration.Service{integration.GoogleSheets, integration.Github}
	for _, s := range userSpecific {
		if s == svc {
			return true
		}
	}
	return false
}

// ParseSetIntegrationAsStorage returns whether the integration being connected should be
// used as the new storage layer.
func ParseSetIntegrationAsStorage(r *http.Request, svc integration.Service) (bool, error) {
	if svc != integration.S3 {
		// Only S3 integrations can be used for storage currently
		return false, nil
	}

	setStorageHeader := r.Header.Get(routes.IntegrationSetStorageHeader)
	setStorage, err := strconv.ParseBool(setStorageHeader)
	if err != nil {
		return false, err
	}

	return setStorage, nil

}
