import { Edge, Node, Position } from 'reactflow';

import {
  ArtifactResponse,
  ArtifactResultResponse,
  NodeResultsMap,
  NodesMap,
  OperatorResponse,
  OperatorResultResponse,
} from '../handlers/responses/node';
import { DagResponse } from '../handlers/responses/workflow';
import { NodesGetResponse } from '../handlers/v2/NodesGet';
import { OperatorType } from '../utils/operators';

type Layer = { ids: string[]; numActiveEdges: number };

export type NodePosition = { x: number; y: number };
export type NodePositions = { [id: string]: NodePosition };
export type VisualizedDag = {
  nodes: Node<ReactFlowNodeData>[];
  edges: Edge[];
};

export type ReactFlowNodeData = {
  targetHandlePosition?: Position;
  nodeType: keyof NodesGetResponse;
  nodeId: string;
  dag: DagResponse;

  // Hack for merged metrics / checks for now.
  // We should be able to use a single node / result
  // once we have dedicated metric / check nodes.
  operator?: OperatorResponse;
  artifact?: ArtifactResponse;
  operatorResult?: OperatorResultResponse;
  artifactResult?: ArtifactResultResponse;
};

const NodeBaseX = 100;
const NodeBaseY = 200;
const IndentX = 500;
const IndentY = 250;

function layerNodes(operators: { [id: string]: OperatorResponse }): Layer[] {
  // start with an empty layer
  const layers: Layer[] = [{ ids: [], numActiveEdges: 0 }];
  const opUpstreamCount: { [id: string]: number } = {};
  let activeEdges = 0;
  const artfToDownstream: { [id: string]: string[] } = {};

  // sort by name
  const opList = Object.values(operators).sort((a, b) =>
    a.name < b.name ? -1 : 1
  );

  // fill the first layer and initialize up / downstream maps
  opList.forEach((op) => {
    op.inputs.forEach((artfId) => {
      if (artfToDownstream[artfId] === undefined) {
        artfToDownstream[artfId] = [];
      }
      artfToDownstream[artfId].push(op.id);
    });
    opUpstreamCount[op.id] = op.inputs.length;
    if (op.inputs.length === 0) {
      layers[0].ids.push(op.id);
    }
  });

  while (layers[layers.length - 1].ids.length > 0) {
    const frontier = layers[layers.length - 1];
    layers.push({ ids: [], numActiveEdges: 0 });
    frontier.ids.forEach((opId) => {
      const op = operators[opId];
      if (!!op) {
        op.outputs.forEach((artfId) => {
          if (artfToDownstream[artfId] !== undefined) {
            artfToDownstream[artfId].forEach((downstreamOpId) => {
              activeEdges += 1;
              opUpstreamCount[downstreamOpId] -= 1;
              if (opUpstreamCount[downstreamOpId] == 0) {
                layers[layers.length - 1].ids.push(downstreamOpId);
                activeEdges -= operators[downstreamOpId].inputs.length;
              }
            });
          }
        });
      }
    });

    layers[layers.length - 1].numActiveEdges = activeEdges;
  }

  return layers;
}

function positionNodes(operators: {
  [id: string]: OperatorResponse;
}): NodePositions {
  const layers = layerNodes(operators);
  const positions: NodePositions = {};

  let opX = NodeBaseX;
  layers.forEach((layer) => {
    const artfX = opX + IndentX;
    let artfY = NodeBaseY + layer.numActiveEdges * IndentY;
    layer.ids.forEach((opId) => {
      positions[opId] = { x: opX, y: artfY };
      const op = operators[opId];
      if (op.outputs.length === 0) {
        // indent 'starting point' for next operator even if this operator has no outputs
        artfY += IndentY;
      }

      op.outputs.forEach((artfId) => {
        positions[artfId] = { x: artfX, y: artfY };
        artfY += IndentY;
      });
    });

    opX = artfX + IndentX;
  });

  return positions;
}

function collapsePosition(
  visualizedDag: VisualizedDag,
  nodes: NodesMap,
  nodeResults?: NodeResultsMap
): VisualizedDag {
  const collapsingOp = Object.values(nodes.operators).filter(
    (op) =>
      op.spec?.type === OperatorType.Check ||
      op.spec?.type === OperatorType.Metric
  );
  const nodeComponents = visualizedDag.nodes;
  const collapsedArtfIds = new Set();

  const nodeComponentsMap: { [id: string]: Node<ReactFlowNodeData> } = {};
  nodeComponents.forEach((n) => {
    nodeComponentsMap[n.id] = n;
  });

  // This map is useful to re-route edges
  const collapsedArtfIdToUpstreamOpId: { [artfId: string]: string } = {};

  collapsingOp.forEach((op) => {
    // we only reduce metric / check nodes if its output size is 1
    if (op.outputs.length === 1) {
      const artfId = op.outputs[0];
      const artfResult = (nodeResults?.artifacts ?? {})[artfId];
      if (!!nodeComponentsMap[op.id]) {
        nodeComponentsMap[op.id].data.artifact = nodes.artifacts[artfId];
        nodeComponentsMap[op.id].data.artifactResult = artfResult;
        collapsedArtfIds.add(artfId);
        collapsedArtfIdToUpstreamOpId[artfId] = op.id;
      }
    }
  });

  const enrichedNodes = Object.values(nodeComponentsMap);
  const filteredNodes = enrichedNodes.filter(
    (node) => !collapsedArtfIds.has(node.id)
  );

  const edges = visualizedDag.edges ?? [];
  // route the edges from removed artf nodes to op nodes they collapsed into.
  edges.forEach((e) => {
    if (collapsedArtfIds.has(e.source)) {
      e.source = collapsedArtfIdToUpstreamOpId[e.source];
    }
  });

  // remove any edge who's target is a metric artifact.
  const filteredEdges = edges
    .filter((e) => !collapsedArtfIds.has(e.target))
    .filter((edge) => {
      // Check if the edge exists in the mappedNodes array.
      // If it does not exist, remove the edge. elk crashes if an edge does not have a corresponding node.
      const nodeFound = false;
      for (let i = 0; i < filteredNodes.length; i++) {
        if (edge.source === filteredNodes[i].id) {
          return true;
        }
      }

      return nodeFound;
    });

  return {
    edges: filteredEdges,
    nodes: filteredNodes,
  };
}

export function getOperatorNode(
  op: OperatorResponse,
  dag: DagResponse,
  pos: NodePosition,
  opResult?: OperatorResultResponse | undefined
): Node<ReactFlowNodeData> {
  return {
    id: op.id,
    type: 'operators',
    draggable: false,
    data: {
      dag,
      nodeType: 'operators',
      nodeId: op.id,
      operator: op,
      operatorResult: opResult,
    },
    position: pos,
  };
}

export function getArtifactNode(
  artf: ArtifactResponse,
  dag: DagResponse,
  pos: NodePosition,
  artfResult?: ArtifactResultResponse | undefined
): Node<ReactFlowNodeData> {
  return {
    id: artf.id,
    type: 'artifacts',
    draggable: false,
    data: {
      dag,
      nodeType: 'artifacts',
      nodeId: artf.id,
      artifact: artf,
      artifactResult: artfResult,
    },
    position: pos,
  };
}

export function getEdges(operators: {
  [id: string]: OperatorResponse;
}): Edge[] {
  const results = [];
  Object.values(operators).forEach((op) => {
    op.inputs.forEach((artfId) => {
      results.push({
        id: `${artfId}-${op.id}`,
        type: 'curved', // Op inputs are curved edges
        source: artfId,
        target: op.id,
        container: 'root',
      });
    });
    op.outputs.forEach((artfId) => {
      results.push({
        id: `${op.id}-${artfId}`,
        type: 'straight',
        source: op.id,
        target: artfId,
        container: 'root',
      });
    });
  });
  return results;
}

export function visualizeDag(
  dag: DagResponse,
  nodes: NodesMap,
  nodeResults?: NodeResultsMap | undefined
): VisualizedDag {
  const nodePositions = positionNodes(nodes.operators);
  const opNodeComponents = Object.values(nodes.operators)
    .filter((op) => op.spec?.type !== OperatorType.Param)
    .map((op) => {
      const pos = nodePositions[op.id];
      if (!pos) {
        return null;
      }

      return getOperatorNode(
        op,
        dag,
        pos,
        (nodeResults?.operators ?? {})[op.id]
      );
    })
    .filter((x) => x !== null);
  const artfNodeComponents = Object.values(nodes.artifacts)
    .map((artf) => {
      const pos = nodePositions[artf.id];
      if (!pos) {
        return null;
      }

      return getArtifactNode(
        artf,
        dag,
        pos,
        (nodeResults?.artifacts ?? {})[artf.id]
      );
    })
    .filter((x) => x !== null);

  const edges = getEdges(nodes.operators);
  const fullVisualizedDag = {
    nodes: opNodeComponents.concat(artfNodeComponents),
    edges,
  };

  return collapsePosition(fullVisualizedDag, nodes, nodeResults);
}
