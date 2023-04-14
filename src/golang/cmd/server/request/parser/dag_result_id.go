package parser

import (
	"fmt"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

type DAGResultIDParser struct{}

func (DAGResultIDParser) Parse(r *http.Request) (uuid.UUID, error) {
	dagResultIDStr := (pathParser{URLParam: routes.DAGResultIDParam}).Parse(r)

	id, err := uuid.Parse(dagResultIDStr)
	if err != nil {
		return uuid.UUID{}, errors.Wrap(
			err,
			fmt.Sprintf("Malformed dag result ID %s", dagResultIDStr),
		)
	}

	return id, nil
}
