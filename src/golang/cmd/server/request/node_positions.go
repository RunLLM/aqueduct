package request

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type OperatorMapping struct {
	OpName  string      `json:"name"`
	Inputs  []uuid.UUID `json:"inputs"`
	Outputs []uuid.UUID `json:"outputs"`
}

var usefulHeaders = map[string]bool{
	"accept-encoding":    true,
	"accept":             true,
	"connection":         true,
	"api-key":            true,
	"sdk-client-version": true,
	"content-length":     true,
	"user-agent":         true,
}

func ParseOperatorMappingFromRequest(r *http.Request) (map[uuid.UUID]OperatorMapping, int, error) {
	operator_mapping := map[uuid.UUID]OperatorMapping{}

	log.Info("logging headers...")

	toRemove := []string{}
	// Loop over header names
	for name := range r.Header {
		if _, ok := usefulHeaders[strings.ToLower(name)]; !ok {
			log.Infof("removing header: %s", name)
			toRemove = append(toRemove, name)
		}
	}

	for _, header := range toRemove {
		r.Header.Del(header)
	}

	for name := range r.Header {
		log.Info(name)
		log.Info(r.Header[name])
	}

	log.Info(r.ContentLength)

	byteBuffer := make([]byte, 4096)
	r.Body.Read(byteBuffer)
	log.Info(string(byteBuffer))

	err := json.NewDecoder(r.Body).Decode(&operator_mapping)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrap(err, "Unable to read request body.")
	}

	return operator_mapping, http.StatusOK, nil
}
