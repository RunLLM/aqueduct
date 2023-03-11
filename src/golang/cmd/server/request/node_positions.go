package request

import (
	"encoding/json"
	"net/http"

	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/google/uuid"
)

type OperatorMapping struct {
	OpName  string      `json:"name"`
	Inputs  []uuid.UUID `json:"inputs"`
	Outputs []uuid.UUID `json:"outputs"`
}

func ParseOperatorMappingFromRequest(r *http.Request) (map[uuid.UUID]OperatorMapping, int, error) {
	operator_mapping := map[uuid.UUID]OperatorMapping{}

	err := json.NewDecoder(r.Body).Decode(&operator_mapping)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Unable to read request body.")
	}

	return operator_mapping, http.StatusOK, nil
}
