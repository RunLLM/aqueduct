package parser

import (
	"fmt"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

type ResourceIDParser struct{}

func (ResourceIDParser) Parse(r *http.Request) (*uuid.UUID, error) {
	resourceIDStr := (pathParser{URLParam: routes.ResourceIDParam}).Parse(r)

	id, err := uuid.Parse(resourceIDStr)
	if err != nil {
		return nil, errors.Wrap(
			err,
			fmt.Sprintf("Malformed resource ID %s", resourceIDStr),
		)
	}

	return &id, nil
}
