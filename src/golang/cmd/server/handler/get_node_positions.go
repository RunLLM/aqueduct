package handler

import (
	"context"
	"net/http"
	"sort"

	"github.com/aqueducthq/aqueduct/cmd/server/request"
	aq_context "github.com/aqueducthq/aqueduct/lib/context"
	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/google/uuid"
)

// Route: /workflow/positioning
// Method: POST
// Params: None
// Request:
//	Headers:
//		`api-key`: user's API Key
//	Body:
//		`operatorIdToInputOutput`: mapping of operator id to the input and output of the operators
// Response:
//	Body:
//		Mapping of operator ids and the node positions (x, y),
//		mapping of artifact ids and the node positions (x, y).

type GetNodePositionsHandler struct {
	PostHandler
}

type getNodePositionsHandlerArgs struct {
	*aq_context.AqContext
	operatorIdToInputOutput map[uuid.UUID]request.OperatorMapping
}

// (x, y) positions for the nodes. First node is at (NodeBaseX=100, NodeBaseY=200) and expands right and down from there.
// (x+i*IndentX=325, y+j*IndentY=300)
type nodePositions struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type getNodePositionsHandlerResponse struct {
	OperatorPositions map[uuid.UUID]nodePositions `json:"operator_positions"`
	ArtifactPositions map[uuid.UUID]nodePositions `json:"artifact_positions"`
}

func (*GetNodePositionsHandler) Name() string {
	return "GetNodePosition"
}

func (*GetNodePositionsHandler) Prepare(r *http.Request) (interface{}, int, error) {
	aqContext, statusCode, err := aq_context.ParseAqContext(r.Context())
	if err != nil {
		return nil, statusCode, err
	}

	operatorIdToInputOutput, statusCode, err := request.ParseOperatorMappingFromRequest(r)
	if err != nil {
		return nil, statusCode, errors.Wrap(err, "Unable to parse operator mapping.")
	}

	return &getNodePositionsHandlerArgs{
		AqContext:               aqContext,
		operatorIdToInputOutput: operatorIdToInputOutput,
	}, http.StatusOK, nil
}

// This is a simplified version of Sugiyama method https://blog.disy.net/sugiyama-method/
// where we compute the DAG topology for rendering its layout.

// It takes a list of operators and returns the following:
//   - Layers, which is a list of layers, where each layer is a list of opId
//   - Number of active edges for each layer. 'active edges' stands for any edges at this layer, that doesn't starts
//     from the previous layer, nor ends at current layer. This is used to compute how much `room` one layer
//     need to save for rendering edges.
//
// The canvas is assumed to be infinite and spacing between nodes are at an arbitrary constant.
// The positions may need to be resized or shifted if the canvas view does not automatically adjust to contain everything.
func orderNodes(operatorIdToInputOutput map[uuid.UUID]request.OperatorMapping) ([][]uuid.UUID, []int) {
	artifactToDownstream := make(map[uuid.UUID][]uuid.UUID)
	upstreamCount := make(map[uuid.UUID]int)
	layers := [][]uuid.UUID{}
	activeLayerEdges := []int{}
	activeEdges := 0
	layers = append(layers, []uuid.UUID{})
	activeLayerEdges = append(activeLayerEdges, 0)

	// make a map with key: op uuid ; val: op name
	opNameIdPair := make(map[string]uuid.UUID)
	opNames := []string{}
	for opId, op := range operatorIdToInputOutput {
		opNameIdPair[op.OpName] = opId
		opNames = append(opNames, op.OpName)
	}
	// sort by name
	sort.Strings(opNames)

	// retrieve a ordering for uuid
	for _, name := range opNames {
		opId := opNameIdPair[name]
		op := operatorIdToInputOutput[opId]
		for _, artfId := range op.Inputs {
			_, ok := artifactToDownstream[artfId]
			if !ok {
				artifactToDownstream[artfId] = []uuid.UUID{}
			}
			artifactToDownstream[artfId] = append(artifactToDownstream[artfId], opId)
		}
		upstreamCount[opId] = len(op.Inputs)
		if len(op.Inputs) == 0 {
			layers[len(layers)-1] = append(layers[len(layers)-1], opId)
		}
	}

	for len(layers[len(layers)-1]) > 0 {
		frontier := layers[len(layers)-1]
		layers = append(layers, []uuid.UUID{})
		for _, opId := range frontier {
			op := operatorIdToInputOutput[opId]
			for _, artfId := range op.Outputs {
				_, ok := artifactToDownstream[artfId]
				if ok {
					for _, downstreamOpId := range artifactToDownstream[artfId] {
						activeEdges += 1
						upstreamCount[downstreamOpId] -= 1
						if upstreamCount[downstreamOpId] == 0 {
							layers[len(layers)-1] = append(layers[len(layers)-1], downstreamOpId)
							activeEdges -= len(operatorIdToInputOutput[downstreamOpId].Inputs)
						}
					}
				}
			}
		}
		activeLayerEdges = append(activeLayerEdges, activeEdges)
	}

	return layers, activeLayerEdges
}

func positionNodes(operators map[uuid.UUID]request.OperatorMapping) (map[uuid.UUID]nodePositions, map[uuid.UUID]nodePositions) {
	NodeBaseX := 100
	NodeBaseY := 200
	IndentX := 325
	IndentY := 300

	layers, activeLayerEdges := orderNodes(operators)

	opPos := map[uuid.UUID]nodePositions{}
	artfPos := map[uuid.UUID]nodePositions{}

	opX := NodeBaseX
	for idx, layer := range layers {
		artfX := opX + IndentX
		artfY := NodeBaseY + activeLayerEdges[idx]*IndentY
		for _, opId := range layer {
			op := operators[opId]
			opPos[opId] = nodePositions{X: opX, Y: artfY}
			if len(op.Outputs) == 0 {
				artfY += IndentY // indent 'starting point' for next operator even if this operator has no outputs
			}
			for _, artfId := range op.Outputs {
				artfPos[artfId] = nodePositions{X: artfX, Y: artfY}
				artfY += IndentY
			}
		}
		opX = artfX + IndentX
	}

	return opPos, artfPos
}

func (*GetNodePositionsHandler) Perform(
	ctx context.Context,
	interfaceArgs interface{},
) (interface{}, int, error) {
	args := interfaceArgs.(*getNodePositionsHandlerArgs)

	opPositions, artPositions := positionNodes(args.operatorIdToInputOutput)

	return getNodePositionsHandlerResponse{
		OperatorPositions: opPositions,
		ArtifactPositions: artPositions,
	}, http.StatusOK, nil
}
