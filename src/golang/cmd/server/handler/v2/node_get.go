package v2

import (
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/request/parser"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/google/uuid"
)

// This is the shared piece for all node get handlers.
type nodeGetArgs struct {
	*aq_context.AqContext
	workflowID uuid.UUID
	dagID      uuid.UUID
	nodeID     uuid.UUID
}

type nodeGetHandler struct{}

func (h *nodeGetHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	workflowID, err := (parser.WorkflowIDParser{}).Parse(r)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	dagID, err := (parser.DagIDParser{}).Parse(r)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	nodeID, err := (parser.NodeIDParser{}).Parse(r)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	return &nodeGetArgs{
		AqContext:  aqContext,
		workflowID: workflowID,
		dagID:      dagID,
		nodeID:     nodeID,
	}, http.StatusOK, nil
}
