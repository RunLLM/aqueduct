package request

import (
	"encoding/json"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/dropbox/godropbox/errors"
)

// ParseIntegrationServiceFromRequest parses the integration service, and whether the
// service is user only.
func ParseIntegrationServiceFromRequest(r *http.Request) (shared.Service, bool, error) {
	serviceStr := r.Header.Get(routes.IntegrationServiceHeader)
	service, err := shared.ParseService(serviceStr)
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
	if len(configHeader) > 0 {
		err := json.Unmarshal([]byte(configHeader), &configuration)
		if err != nil {
			return "", nil, errors.Wrap(err, "Unable to parse integration configuration: %v")
		}
	}

	integrationName := r.Header.Get(routes.IntegrationNameHeader)
	if integrationName == "" {
		return "", nil, errors.New("Integration name was not provided.")
	}

	return integrationName, configuration, nil
}

// isUserOnlyIntegration returns whether the specified service is only accessible by the user.
func isUserOnlyIntegration(svc shared.Service) bool {
	userSpecific := []shared.Service{
		shared.GoogleSheets,
		shared.Github,
		shared.Conda,
		shared.Email,
		shared.Slack,
	}
	for _, s := range userSpecific {
		if s == svc {
			return true
		}
	}
	return false
}
