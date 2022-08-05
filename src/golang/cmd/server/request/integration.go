package request

import (
	"encoding/json"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	"github.com/dropbox/godropbox/errors"
)

var ErrNoIntegrationName = errors.New("Integration name is not provided")

// ParseIntegrationServiceFromRequest parses the integration service, and whether the
// service is user only.
func ParseIntegrationServiceFromRequest(r *http.Request) (integration.Service, bool, error) {
	serviceStr := r.Header.Get(routes.IntegrationServiceHeader)
	service, err := integration.ParseService(serviceStr)
	if err != nil {
		return "", false, err
	}

	return service, isUserOnlyIntegration(service), nil
}

// ParseIntegrationConfigFromRequest parses the integration name and configuration,
// from the request
func ParseIntegrationConfigFromRequest(r *http.Request) (string, map[string]string, error) {

	configHeader := r.Header.Get(routes.IntegrationConfigHeader)
	var configuration map[string]string
	err := json.Unmarshal([]byte(configHeader), &configuration)
	if err != nil {
		return "", nil, errors.Wrap(err, "Unable to parse integration configuration: %v")
	}

	integrationName := r.Header.Get(routes.IntegrationNameHeader)
	if integrationName == "" {
		return "", nil, ErrNoIntegrationName
	}

	return integrationName, configuration, nil
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
