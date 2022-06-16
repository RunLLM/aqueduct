package request

import (
	"encoding/json"
	"net/http"
)

const (
	// The value is a json-stringified dictionary of parameter names to values.
	parametersKey = "parameters"
)

// Extracts the parameters dictionary from the request, which maps
// parameter name to its json-serialized string value.
func ExtractParamsfromRequest(r *http.Request) (map[string]string, error) {
	serializedParams := []byte(r.FormValue(parametersKey))

	// No-op if there aren't any parameters set.
	if len(serializedParams) == 0 {
		return nil, nil
	}

	var params map[string]string
	err := json.Unmarshal(serializedParams, &params)
	if err != nil {
		return nil, err
	}
	return params, nil
}
