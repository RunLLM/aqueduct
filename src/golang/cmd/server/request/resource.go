package request

import (
	"encoding/json"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/dropbox/godropbox/errors"
)

// ParseResourceServiceFromRequest parses the resource service, and whether the
// service is user only.
func ParseResourceServiceFromRequest(r *http.Request) (shared.Service, bool, error) {
	serviceStr := r.Header.Get(routes.ResourceServiceHeader)
	service, err := shared.ParseService(serviceStr)
	if err != nil {
		return "", false, err
	}

	return service, isUserOnlyResource(service), nil
}

// ParseResourceConfigFromRequest parses the resource name and configuration,
// from the request
func ParseResourceConfigFromRequest(r *http.Request) (string, map[string]string, error) {
	configHeader := r.Header.Get(routes.ResourceConfigHeader)
	var configuration map[string]string
	if len(configHeader) > 0 {
		err := json.Unmarshal([]byte(configHeader), &configuration)
		if err != nil {
			return "", nil, errors.Wrap(err, "Unable to parse resource configuration: %v")
		}
	}

	resourceName := r.Header.Get(routes.ResourceNameHeader)
	if resourceName == "" {
		return "", nil, errors.New("Resource name was not provided.")
	}

	return resourceName, configuration, nil
}

// isUserOnlyResource returns whether the specified service is only accessible by the user.
func isUserOnlyResource(svc shared.Service) bool {
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
