package request

import (
	"encoding/json"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/lib/collections/integration"
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

	userOnly := integration.IsUserOnlyIntegration(service)

	return service, integrationName, configuration, userOnly, nil
}
