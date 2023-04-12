package parser

import (
	"fmt"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/routes"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

type WorkflowIDParser struct{}

func (WorkflowIDParser) Parse(r *http.Request) (uuid.UUID, error) {
	workflowIDStr := (pathParser{URLParam: routes.WorkflowIDParam}).Parse(r)

	id, err := uuid.Parse(workflowIDStr)
	if err != nil {
		return uuid.UUID{}, errors.Wrap(
			err,
			fmt.Sprintf("Malformed workflow ID %s", workflowIDStr),
		)
	}

	return id, nil
}
