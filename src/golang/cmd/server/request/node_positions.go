package request

import (
	"encoding/json"
	"net/http"

	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type OperatorMapping struct {
	OpName  string      `json:"name"`
	Inputs  []uuid.UUID `json:"inputs"`
	Outputs []uuid.UUID `json:"outputs"`
}

func ParseOperatorMappingFromRequest(r *http.Request) (map[uuid.UUID]OperatorMapping, int, error) {
	operator_mapping := map[uuid.UUID]OperatorMapping{}

	log.Info("logging headers...")
	for name := range r.Header {
		log.Info(name)
		log.Info(r.Header[name])
	}

	log.Info(r.ContentLength)

	err := json.NewDecoder(r.Body).Decode(&operator_mapping)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Unable to read request body.")
	}

	return operator_mapping, http.StatusOK, nil
}
