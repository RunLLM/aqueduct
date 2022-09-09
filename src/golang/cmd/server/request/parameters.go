package request

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/dropbox/godropbox/errors"
)

const (
	// The value is a json-stringified dictionary of parameter names to values.
	parametersKey = "parameters"
)

// Extracts the parameters dictionary from the request, which maps
// parameter name to its base64-encoded, bytes-serialized value.
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

	// Check that the parameter string is base-64 encoded.
	for param_name, param_val := range params {
		_, err = base64.StdEncoding.DecodeString(param_val)
		if err != nil {
			return nil, errors.Newf("Internal error: parameter values must be base64-encoded. The value %s provided for parameter %s was not.", param_val, param_name)
		}
	}
	return params, nil
}
