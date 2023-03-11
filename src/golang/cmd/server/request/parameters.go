package request

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/aqueducthq/aqueduct/lib/models/shared/operator/param"
)

// Requests that include overwriting parameters are expected to have this json format in the body.
//type expectedRequestBody struct {
//	SerializedParams map[string]string `json:"parameters"`
//}

const (
	// The value is a json'ed dictionary of parameter names to another
	// dictionary to the parameter's base64 encoded value and serialization type.
	parametersKey = "parameters"

	paramValKey               = "val"
	paramSerializationTypeKey = "serialization_type"
)

// Extracts the parameters dictionary from the request, which maps
// parameter name to its base64-encoded value and serialization type.
func ExtractParamsfromRequest(r *http.Request) (map[string]param.Param, error) {
	serializedParams := []byte(r.FormValue(parametersKey))

	// No-op if there aren't any parameters set.
	if len(serializedParams) == 0 {
		return nil, nil
	}

	var paramSpecMapByName map[string]map[string]string
	err := json.Unmarshal(serializedParams, &paramSpecMapByName)
	if err != nil {
		return nil, err
	}

	// Convert each value into a param spec.
	paramSpecByName := make(map[string]param.Param, len(paramSpecMapByName))
	for paramName, paramSpecMap := range paramSpecMapByName {
		encodedParamVal := paramSpecMap[paramValKey]
		serializationType := paramSpecMap[paramSerializationTypeKey]

		// Check that the parameter value is base64 encoded.
		_, err = base64.StdEncoding.DecodeString(encodedParamVal)
		if err != nil {
			return nil, errors.Newf(
				"Internal error: parameter values must be base64-encoded. "+
					"The value %s provided for parameter %s was not.", encodedParamVal, paramName)
		}

		paramSpecByName[paramName] = param.Param{
			Val:               encodedParamVal,
			SerializationType: serializationType,
		}
	}

	return paramSpecByName, nil
}
