package parser

import (
	"fmt"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

type IntegrationIDParser struct{}

func (IntegrationIDParser) Parse(r *http.Request) (*uuid.UUID, error) {
	integrationIDStr := (pathParser{URLParam: routes.IntegrationIDParam}).Parse(r)

	id, err := uuid.Parse(integrationIDStr)
	if err != nil {
		return nil, errors.Wrap(
			err,
			fmt.Sprintf("Malformed integration ID %s", integrationIDStr),
		)
	}

	return &id, nil
}
