package request

import (
	"encoding/base64"
	"encoding/json"
	"github.com/aqueducthq/aqueduct/lib/collections/operator/param"
	"github.com/dropbox/godropbox/errors"
	log "github.com/sirupsen/logrus"
	"net/http"
)

// Requests that include overwriting parameters are expected to have this json format in the body.
type expectedRequestBody struct {
	SerializedParams map[string]string `json:"parameters"`
}

// Extracts the parameters dictionary from the request, which maps
// parameter name to its base64-encoded value and serialization type.
func ExtractParamsfromRequest(r *http.Request) (map[string]param.Param, error) {
	var body expectedRequestBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse JSON body.")
	}

	log.Errorf("HELLO: request body: ", body)

	serializedParams := body.SerializedParams

	// No-op if there aren't any parameters set.
	if len(serializedParams) == 0 {
		return nil, nil
	}

	// Because each parameter spec was individually json-serialized, we have to go through and
	// decode each one.
	params := make(map[string]param.Param, len(serializedParams))
	for paramName, serializedParam := range serializedParams {
		var paramSpec param.Param
		err = json.Unmarshal([]byte(serializedParam), &paramSpec)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to parse parameter value.")
		}

		// Check that the parameter value is base64 encoded.
		_, err = base64.StdEncoding.DecodeString(paramSpec.Val)
		if err != nil {
			return nil, errors.Newf(
				"Internal error: parameter values must be base64-encoded. "+
					"The value %s provided for parameter %s was not.", paramSpec.Val, paramName)
		}

		params[paramName] = paramSpec
	}

	return params, nil
}
