package parser

import (
	"fmt"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

type NodeIDParser struct{}

func (NodeIDParser) Parse(r *http.Request) (uuid.UUID, error) {
	nodeIDStr := (pathParser{URLParam: routes.NodeIDParam}).Parse(r)

	id, err := uuid.Parse(nodeIDStr)
	if err != nil {
		return uuid.UUID{}, errors.Wrap(
			err,
			fmt.Sprintf("Malformed node ID %s", nodeIDStr),
		)
	}

	return id, nil
}
