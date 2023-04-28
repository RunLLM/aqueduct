package parser

import (
	"github.com/google/uuid"
	"net/http"
)

type IntegrationIDQueryParser struct{}

func (IntegrationIDQueryParser) Parse(r *http.Request) (uuid.UUID, error) {
	query := r.URL.Query()

	integrationID, err := uuid.Parse(query.Get("integrationID"))
	if err != nil {
		return uuid.UUID{}, err
	}
	return integrationID, nil
}
