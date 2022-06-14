package request

import (
	"encoding/json"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
)

const (
	// The value is a json-stringified dictionary of parameter names to values.
	parametersKey = "parameters"
)

// Extracts the parameters dictionary from the request, which maps
// parameter name to its json-serialized string value.
func ExtractParamsfromRequest(r *http.Request) (map[string]string, error) {
	paramsInBytes, err := ExtractHttpPayload(
		r.Header.Get(routes.ContentTypeHeader),
		parametersKey,
		false,
		r,
	)
	if err != nil {
		return nil, err
	}

	// No-op if there aren't any parameters set.
	if len(paramsInBytes) == 0 {
		return nil, nil
	}

	var params map[string]string
	err = json.Unmarshal(paramsInBytes, &params)
	if err != nil {
		return nil, err
	}
	return params, nil
}
