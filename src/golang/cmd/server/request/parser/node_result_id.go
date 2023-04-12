package parser

import (
	"fmt"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

type NodeResultIDParser struct{}

func (NodeResultIDParser) Parse(r *http.Request) (uuid.UUID, error) {
	nodeResultIDStr := (pathParser{URLParam: routes.NodeResultIDParam}).Parse(r)

	id, err := uuid.Parse(nodeResultIDStr)
	if err != nil {
		return uuid.UUID{}, errors.Wrap(
			err,
			fmt.Sprintf("Malformed node result ID %s", nodeResultIDStr),
		)
	}

	return id, nil
}
