package parser

import (
	"fmt"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

type DagIDParser struct{}

func (DagIDParser) Parse(r *http.Request) (uuid.UUID, error) {
	dagIDStr := (pathParser{URLParam: routes.DagIDParam}).Parse(r)

	id, err := uuid.Parse(dagIDStr)
	if err != nil {
		return uuid.UUID{}, errors.Wrap(
			err,
			fmt.Sprintf("Malformed DAG ID %s", dagIDStr),
		)
	}

	return id, nil
}
