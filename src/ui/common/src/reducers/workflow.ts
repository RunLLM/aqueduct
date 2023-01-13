// This is being deprecated. Please use `workflowDagResults` combining with
// `artifactResults` for future development.
import { createAsyncThunk, createSlice, PayloadAction } from '@reduxjs/toolkit';
import { Edge, Node } from 'react-flow-renderer';
import ELK from 'elkjs/lib/elk.bundled.js'

import { apiAddress } from '../components/hooks/useAqueductConsts';
import { RootState } from '../stores/store';
import {
  Artifact,
  GetArtifactResultResponse,
  SerializationType,
} from '../utils/artifacts';
import {
  GetOperatorResultResponse,
  Operator,
  OperatorType,
} from '../utils/operators';
import {
  getArtifactNode,
  getEdges,
  getOperatorNode,
  ReactFlowNodeData,
} from '../utils/reactflow';
import { isSucceeded, LoadingStatus, LoadingStatusEnum } from '../utils/shared';
import { ExecutionStatus } from '../utils/shared';
import {
  DeleteWorkflowResponse,
  getSavedObjectIdentifier,
  GetWorkflowResponse,
  ListWorkflowSavedObjectsResponse,
  normalizeGetWorkflowResponse,
  SavedObject,
  SavedObjectDeletion,
  WorkflowDag,
  WorkflowDagResultSummary,
} from '../utils/workflows';

type positionResponse = {
  nodes: Node<ReactFlowNodeData>[];
  edges: Edge[];
};

type selectDagPositionResult = {
  loadingStatus: LoadingStatus;
  result?: positionResponse;
};

export type ArtifactResult = {
  loadingStatus: LoadingStatus;
  result?: GetArtifactResultResponse;
};

export type OperatorResult = {
  loadingStatus: LoadingStatus;
  result?: GetOperatorResultResponse;
};

export type SavedObjectResult = {
  loadingStatus: LoadingStatus;
  result: Record<string, SavedObject[]>;
};

export type SavedObjectDeletionResult = {
  loadingStatus: LoadingStatus;
  result: Record<string, SavedObjectDeletion[]>;
};

export type WorkflowState = {
  loadingStatus: LoadingStatus;
  savedObjects: SavedObjectResult;
  savedObjectDeletion: SavedObjectDeletionResult;
  dags: { [id: string]: WorkflowDag };
  dagResults: WorkflowDagResultSummary[];

  selectedResult?: WorkflowDagResultSummary;
  selectedDag?: WorkflowDag;
  selectedDagPosition?: selectDagPositionResult;
  artifactResults: { [id: string]: ArtifactResult };
  operatorResults: { [id: string]: OperatorResult };
};

const initialState: WorkflowState = {
  loadingStatus: { loading: LoadingStatusEnum.Initial, err: '' },
  savedObjects: {
    loadingStatus: { loading: LoadingStatusEnum.Initial, err: '' },
    result: {},
  },
  savedObjectDeletion: {
    loadingStatus: { loading: LoadingStatusEnum.Initial, err: '' },
    result: {},
  },
  dags: {},
  dagResults: [],
  artifactResults: {},
  operatorResults: {},
  selectedDag: undefined,
  selectedResult: undefined,
  selectedDagPosition: {
    loadingStatus: { loading: LoadingStatusEnum.Initial, err: '' },
    result: { nodes: [], edges: [] },
  },
};

export const handleGetOperatorResults = createAsyncThunk<
  GetOperatorResultResponse,
  { apiKey: string; workflowDagResultId: string; operatorId: string }
>(
  'workflowReducer/getOperatorResults',
  async (
    args: {
      apiKey: string;
      workflowDagResultId: string;
      operatorId: string;
    },
    thunkAPI
  ) => {
    const { apiKey, workflowDagResultId, operatorId } = args;
    const res = await fetch(
      `${apiAddress}/api/operator/${workflowDagResultId}/${operatorId}/result`,
      {
        method: 'GET',
        headers: {
          'api-key': apiKey,
        },
      }
    );

    const body = await res.json();
    if (!res.ok) {
      return thunkAPI.rejectWithValue(body.error);
    }
    return body as GetOperatorResultResponse;
  }
);

export const handleGetArtifactResults = createAsyncThunk<
  GetArtifactResultResponse,
  { apiKey: string; workflowDagResultId: string; artifactId: string }
>(
  'workflowReducer/getArtifactResults',
  async (
    args: {
      apiKey: string;
      workflowDagResultId: string;
      artifactId: string;
    },
    thunkAPI
  ) => {
    const { apiKey, workflowDagResultId, artifactId } = args;
    const res = await fetch(
      `${apiAddress}/api/artifact/${workflowDagResultId}/${artifactId}/result`,
      {
        method: 'GET',
        headers: {
          'api-key': apiKey,
        },
      }
    );

    try {
      const formData = await res.formData();
      const metadataJson = await (formData.get('metadata') as File).text();
      const artifactResult = JSON.parse(
        metadataJson
      ) as GetArtifactResultResponse;

      if (artifactResult.exec_state.status === ExecutionStatus.Succeeded) {
        if (
          artifactResult.serialization_type === SerializationType.String ||
          artifactResult.serialization_type === SerializationType.Table ||
          artifactResult.serialization_type === SerializationType.BsonTable ||
          artifactResult.serialization_type === SerializationType.Json
        ) {
          artifactResult.data = await (formData.get('data') as File).text();
        } else if (
          artifactResult.serialization_type === SerializationType.Image
        ) {
          const toBase64 = (file) =>
            new Promise<string>((resolve, reject) => {
              const reader = new FileReader();
              reader.readAsDataURL(file);
              reader.onload = () =>
                resolve(
                  // Use a regex to remove data url part
                  (reader.result as string)
                    .replace('data:', '')
                    .replace(/^.+,/, '')
                );
              reader.onerror = (error) => reject(error);
            });

          artifactResult.data = await toBase64(formData.get('data') as File);
        }
      }

      return artifactResult;
    } catch (err) {
      return thunkAPI.rejectWithValue(err);
    }
  }
);

export const handleGetWorkflow = createAsyncThunk<
  GetWorkflowResponse,
  { apiKey: string; workflowId: string }
>(
  'workflowReducer/get',
  async (
    args: {
      apiKey: string;
      workflowId: string;
    },
    thunkAPI
  ) => {
    const { apiKey, workflowId } = args;

    const res = await fetch(`${apiAddress}/api/workflow/${workflowId}`, {
      method: 'GET',
      headers: {
        'api-key': apiKey,
      },
    });

    const body = await res.json();
    if (!res.ok) {
      return thunkAPI.rejectWithValue(body.error);
    }

    return normalizeGetWorkflowResponse(body);
  }
);

export const handleListWorkflowSavedObjects = createAsyncThunk<
  ListWorkflowSavedObjectsResponse,
  { apiKey: string; workflowId: string }
>(
  'workflowReducer/getObjects',
  async (
    args: {
      apiKey: string;
      workflowId: string;
    },
    thunkAPI
  ) => {
    const { apiKey, workflowId } = args;

    const res = await fetch(
      `${apiAddress}/api/workflow/${workflowId}/objects`,
      {
        method: 'GET',
        headers: {
          'api-key': apiKey,
        },
      }
    );

    const body = await res.json();
    if (!res.ok) {
      return thunkAPI.rejectWithValue(body.error);
    }

    return body as ListWorkflowSavedObjectsResponse;
  }
);

export const handleDeleteWorkflow = createAsyncThunk<
  DeleteWorkflowResponse,
  { apiKey: string; workflowId: string; selectedObjects: Set<SavedObject> }
>(
  'workflowReducer/deleteWorkflow',
  async (
    args: {
      apiKey: string;
      workflowId: string;
      selectedObjects: Set<SavedObject>;
    },
    thunkAPI
  ) => {
    const { apiKey, workflowId, selectedObjects } = args;

    const data = { force: true };
    data['external_delete'] = {};

    selectedObjects.forEach((object) => {
      if (data['external_delete'][object.integration_name]) {
        data['external_delete'][object.integration_name].push(
          JSON.stringify(object.spec)
        );
      } else {
        data['external_delete'][object.integration_name] = [
          JSON.stringify(object.spec),
        ];
      }
    });

    const res = await fetch(`${apiAddress}/api/workflow/${workflowId}/delete`, {
      method: 'POST',
      headers: {
        'api-key': apiKey,
      },
      body: JSON.stringify(data),
    });

    const body = await res.json();
    if (!res.ok) {
      return thunkAPI.rejectWithValue(body.error);
    }

    return body as DeleteWorkflowResponse;
  }
);

// Update `nodes` in-place and returns whether the update succeeded.
function updateNodeWithResult(
  nodes: Node<ReactFlowNodeData>[],
  id: string,
  data: string
): boolean {
  const matchingIdx = nodes
    .map((node, idx) => (node.id === id ? idx : undefined))
    .filter((x) => x !== undefined);
  if (matchingIdx.length === 1) {
    nodes[matchingIdx[0]].data.result = data;
    return true;
  }

  return false;
}

// This function updates nodes in-place and should be called inside a reducer.
// It updates the node matching artfId with the data.
// If the node doesn't exist, we assume the node is collapsed and we will search
// for the right 'parent' to add.
function updateNodeWithArtifactResult(
  nodes: Node<ReactFlowNodeData>[],
  dag: WorkflowDag,
  artfId: string,
  data: string
): void {
  // try to update using artfId
  if (updateNodeWithResult(nodes, artfId, data)) {
    return;
  }

  // Ndoe with artfId doesn't exist. Figure opId and update with that:
  const matchingOp = Object.values(dag.operators).filter((op) =>
    op.outputs.includes(artfId)
  );
  if (matchingOp.length === 1) {
    const opId = matchingOp[0].id;
    updateNodeWithResult(nodes, opId, data);
  }

  // no match found. For now, we don't handle this case.
}

function collapsePosition(
  position: positionResponse,
  dag: WorkflowDag,
  artifactResults: { [id: string]: ArtifactResult }
): positionResponse {
  const collapsingOp = Object.values(dag.operators).filter(
    (op) =>
      op.spec.type === OperatorType.Check ||
      op.spec.type === OperatorType.Metric
  );
  const nodes = position.nodes;
  const collapsedArtfIds = new Set();

  const nodesMap: { [id: string]: Node<ReactFlowNodeData> } = {};
  nodes.forEach((n) => {
    nodesMap[n.id] = n;
  });

  // This map is useful to re-route edges
  const collapsedArtfIdToUpstreamOpId: { [artfId: string]: string } = {};

  collapsingOp.forEach((op) => {
    // we only reduce metric / check nodes if its output size is 1
    if (op.outputs.length === 1) {
      const artfId = op.outputs[0];
      const artfData = artifactResults[artfId]?.result?.data;
      if (!!nodesMap[op.id]) {
        nodesMap[op.id].data.result = artfData;
        collapsedArtfIds.add(artfId);
        collapsedArtfIdToUpstreamOpId[artfId] = op.id;
      }
    }
  });

  const enrichedNodes = Object.values(nodesMap);
  const filteredNodes = enrichedNodes.filter(
    (node) => !collapsedArtfIds.has(node.id)
  );

  const edges = position.edges ?? [];
  // route the edges from removed artf nodes to op nodes they collapsed into.
  edges.forEach((e) => {
    if (collapsedArtfIds.has(e.source)) {
      e.source = collapsedArtfIdToUpstreamOpId[e.source];
    }
  });

  // remove any edge who's target is a metric artifact.
  const filteredEdges = edges.filter((e) => !collapsedArtfIds.has(e.target));

  return {
    edges: filteredEdges,
    nodes: filteredNodes,
  };
}

export const handleGetSelectDagPosition = createAsyncThunk<
  positionResponse,
  {
    apiKey: string;
    operators: { [id: string]: Operator };
    artifacts: { [id: string]: Artifact };
  },
  { state: RootState }
>(
  'workflowReducer/getSelectDagPosition',
  async (
    args: {
      apiKey: string;
      operators: { [id: string]: Operator };
      artifacts: { [id: string]: Artifact };
      onChange: () => void;
      onConnect: (any) => void;
    },
    thunkAPI
  ) => {
    const { apiKey, operators, artifacts, onChange, onConnect } = args;
    const res = await fetch(`${apiAddress}/api/positioning`, {
      method: 'POST',
      headers: {
        'api-key': apiKey,
      },
      body: JSON.stringify(operators),
    });

    const position = await res.json();
    if (!res.ok) {
      return thunkAPI.rejectWithValue(position.error);
    }

    const opPositions = position.operator_positions;
    const artfPositions = position.artifact_positions;
    const opNodes = Object.values(operators)
      .filter((op) => {
        return op.spec.type != OperatorType.Param;
      })
      .map((op) =>
        getOperatorNode(op, opPositions[op.id], onChange, onConnect)
      );
    const artfNodes = Object.values(artifacts).map((artf) =>
      getArtifactNode(artf, artfPositions[artf.id], onChange, onConnect)
    );
    const edges = getEdges(operators);
    const allNodes = {
      nodes: opNodes.concat(artfNodes),
      edges: edges,
    } as positionResponse;
    const dag = thunkAPI.getState().workflowReducer.selectedDag;
    const artifactResults =
      thunkAPI.getState().workflowReducer.artifactResults ?? {};


    /**
     * [
    {
        "id": "02ee9d4b-6c58-4b20-805c-cb00b792424b",
        "type": "checkOp",
        "draggable": false,
        "data": {
            "nodeType": "operator",
            "nodeId": "02ee9d4b-6c58-4b20-805c-cb00b792424b",
            "label": "valid_probabilities"
        },
        "position": {
            "x": 2700,
            "y": 800
        }
    },
    {
        "id": "0de1a746-99fc-4378-a6f8-2239c9d80687",
        "type": "functionOp",
        "draggable": false,
        "data": {
            "nodeType": "operator",
            "nodeId": "0de1a746-99fc-4378-a6f8-2239c9d80687",
            "label": "predict_tree"
        },
        "position": {
            "x": 1400,
            "y": 800
        }
    },
    {
        "id": "150a8d68-e508-4acc-b28a-797e595e7e5a",
        "type": "checkOp",
        "draggable": false,
        "data": {
            "nodeType": "operator",
            "nodeId": "150a8d68-e508-4acc-b28a-797e595e7e5a",
            "label": "greater than 0.1"
        },
        "position": {
            "x": 3350,
            "y": 200
        }
    },
    {
        "id": "2a6e97c2-fd0d-4aab-bed3-56fd33b919e7",
        "type": "checkOp",
        "draggable": false,
        "data": {
            "nodeType": "operator",
            "nodeId": "2a6e97c2-fd0d-4aab-bed3-56fd33b919e7",
            "label": "less than 0.4"
        },
        "position": {
            "x": 3350,
            "y": 800
        }
    },
    {
        "id": "3ef0268c-07e9-4673-863a-04cbd1e45e68",
        "type": "extractOp",
        "draggable": false,
        "data": {
            "nodeType": "operator",
            "nodeId": "3ef0268c-07e9-4673-863a-04cbd1e45e68",
            "label": "aqueduct_demo query 4"
        },
        "position": {
            "x": 100,
            "y": 200
        }
    },
    {
        "id": "4b02c8bb-15ed-4d3c-9cc4-e20b4e216bdc",
        "type": "checkOp",
        "draggable": false,
        "data": {
            "nodeType": "operator",
            "nodeId": "4b02c8bb-15ed-4d3c-9cc4-e20b4e216bdc",
            "label": "less than 0.3"
        },
        "position": {
            "x": 3350,
            "y": 500
        }
    },
    {
        "id": "53ef91e2-9776-446b-8f11-a2af9ffc83cb",
        "type": "functionOp",
        "draggable": false,
        "data": {
            "nodeType": "operator",
            "nodeId": "53ef91e2-9776-446b-8f11-a2af9ffc83cb",
            "label": "predict_ensemble"
        },
        "position": {
            "x": 2050,
            "y": 200
        }
    },
    {
        "id": "a946fc2a-9c6e-4364-8284-68eae3728000",
        "type": "loadOp",
        "draggable": false,
        "data": {
            "nodeType": "operator",
            "nodeId": "a946fc2a-9c6e-4364-8284-68eae3728000",
            "label": "save to aqueduct_demo 1"
        },
        "position": {
            "x": 2700,
            "y": 500
        }
    },
    {
        "id": "c8402fd5-d09d-4c18-8eec-9ebb0faeb7f8",
        "type": "functionOp",
        "draggable": false,
        "data": {
            "nodeType": "operator",
            "nodeId": "c8402fd5-d09d-4c18-8eec-9ebb0faeb7f8",
            "label": "predict_linear"
        },
        "position": {
            "x": 1400,
            "y": 500
        }
    },
    {
        "id": "eb244f91-424f-4309-8e14-00fc952aa014",
        "type": "metricOp",
        "draggable": false,
        "data": {
            "nodeType": "operator",
            "nodeId": "eb244f91-424f-4309-8e14-00fc952aa014",
            "label": "mean(prob_churn)"
        },
        "position": {
            "x": 2700,
            "y": 200
        }
    },
    {
        "id": "f1fc9258-d843-42f1-94c0-a1b669547770",
        "type": "functionOp",
        "draggable": false,
        "data": {
            "nodeType": "operator",
            "nodeId": "f1fc9258-d843-42f1-94c0-a1b669547770",
            "label": "log_featurize"
        },
        "position": {
            "x": 750,
            "y": 500
        }
    },
    {
        "id": "02a39695-633e-4466-9d49-0a8881a8bca5",
        "type": "tableArtifact",
        "draggable": false,
        "data": {
            "nodeType": "artifact",
            "nodeId": "02a39695-633e-4466-9d49-0a8881a8bca5",
            "label": "predict_tree artifact"
        },
        "position": {
            "x": 1725,
            "y": 800
        }
    },
    {
        "id": "03d47e76-7c7c-4b4a-b572-19508fd6ae9b",
        "type": "tableArtifact",
        "draggable": false,
        "data": {
            "nodeType": "artifact",
            "nodeId": "03d47e76-7c7c-4b4a-b572-19508fd6ae9b",
            "label": "aqueduct_demo query 4 artifact"
        },
        "position": {
            "x": 425,
            "y": 200
        }
    },
    {
        "id": "5da51a9d-5e45-48ef-ac9d-00fcf82834f4",
        "type": "tableArtifact",
        "draggable": false,
        "data": {
            "nodeType": "artifact",
            "nodeId": "5da51a9d-5e45-48ef-ac9d-00fcf82834f4",
            "label": "predict_ensemble artifact"
        },
        "position": {
            "x": 2375,
            "y": 200
        }
    },
    {
        "id": "bc84232e-ad21-4cfd-8e23-4bf8a1e326e3",
        "type": "tableArtifact",
        "draggable": false,
        "data": {
            "nodeType": "artifact",
            "nodeId": "bc84232e-ad21-4cfd-8e23-4bf8a1e326e3",
            "label": "predict_linear artifact"
        },
        "position": {
            "x": 1725,
            "y": 500
        }
    },
    {
        "id": "f841bba0-384e-4f36-9904-9ff0276112ec",
        "type": "tableArtifact",
        "draggable": false,
        "data": {
            "nodeType": "artifact",
            "nodeId": "f841bba0-384e-4f36-9904-9ff0276112ec",
            "label": "log_featurize artifact"
        },
        "position": {
            "x": 1075,
            "y": 500
        }
    }
]
     */

    if (!!dag) {
      let collapsedPosition = collapsePosition(allNodes, dag, artifactResults);
      console.log('getSelectDagPosition collapsedPosition: ', collapsedPosition);

      // Do the stuff with the DAG here...
      console.log('initializing ELK');
      const elk = new ELK();

      // const graph = {
      //   id: "root",
      //   layoutOptions: { 'elk.algorithm': 'layered' },
      //   children: [
      //     //{ id: "n1", width: 30, height: 30 },
      //     { id: "n1", width: 30, height: 30 },
      //     { id: "n2", width: 30, height: 30 },
      //     { id: "n3", width: 30, height: 30 }
      //   ],
      //   edges: [
      //     { id: "e1", sources: ["n1"], targets: ["n2"] },
      //     { id: "e2", sources: ["n1"], targets: ["n3"] }
      //   ]
      // }

      const mappedNodes = collapsedPosition.nodes.map((node) => {
        return {
          id: node.id,
          width: 250,
          height: 250,
        }
      });

      console.log('mappedNodes: ', mappedNodes);

      const mappedEdges = collapsedPosition.edges.map((edge) => {
        return {
          id: edge.id,
          sources: [edge.source],
          targets: [edge.target]
        }
      }).filter((mappedEdge) => {
        // Check if the edge exists in the mappedNodes array.
        // If it does not exist, remove the edge. elk crashes if an edge does not have a corresponding node.
        let nodeFound = false;
        for (let i = 0; i < mappedNodes.length; i++) {
          if (mappedEdge.sources[0] === mappedNodes[i].id) {
            console.log('found node in sources');
            return true;
          } else if (mappedEdge.targets[0] === mappedNodes[i].id) {
            console.log('found node in target')
            return true;
          }
        }

        console.log('returning nodeFound val: ', nodeFound);
        return nodeFound;
      })

      console.log('mappedEdges: ', mappedEdges);

      const graph = {
        id: "root",
        layoutOptions: { 'elk.algorithm': 'layered' },
        children: mappedNodes,
        edges: mappedEdges
      }

      const positionedLayout = await elk.layout(graph);
      console.log('positionedLayout: ', positionedLayout);

      collapsedPosition.nodes = collapsedPosition.nodes.map((node) => {
        console.log('processedNodes node: ', node);
        console.log('positionedLayout children length: ', positionedLayout.children.length);
        for (let i = 0; i < positionedLayout.children.length; i++) {
          console.log('nodeId: ', node.id);
          console.log('positionedLayout.children[i].id: ', positionedLayout.children[i].id);
          if (node.id === positionedLayout.children[i].id) {
            console.log('node found setting layout ...');
            node.position = {
              x: positionedLayout.children[i].x,
              y: positionedLayout.children[i].y
            }
          }
        }

        return node;
      })

      //console.log('processedNodes result: ', processedNodes);

      // collapsedPosition.nodes = collapsedPosition.nodes.map((node) => {
      //   console.log('positionedLayout children length: ', positionedLayout.children.length);
      //   debugger;
      //   for (let i = 0; i < positionedLayout.children.length; i++) {
      //     console.log('nodeId: ', node.id);
      //     console.log('positionedLayout.children[i].id: ', positionedLayout.children[i].id);
      //     if (node.id === positionedLayout.children[i].id) {
      //       console.log('node found setting layout ...')
      //       node.position = {
      //         x: positionedLayout[i].x,
      //         y: positionedLayout[i].y
      //       }
      //     }
      //   }

      //   return node;
      // })

      console.log('collapsedPosition after elk: ', collapsedPosition);

      // elk.layout(graph)
      //   .then(console.log)
      //   .catch(console.error)

      return collapsedPosition;
    }

    console.log('getSelectDagPosition allNodes: ', allNodes);
    return allNodes;
  }
);

export const workflowSlice = createSlice({
  name: 'workflowReducer',
  initialState,
  reducers: {
    resetState: () => initialState,
    selectResultIdx: (state, { payload }: PayloadAction<number>) => {
      state.artifactResults = {};
      state.operatorResults = {};

      state.selectedResult = state.dagResults[payload];
      // check if we have a currently selectedResult. If not, then set to a value like 0 so that we don't cause an error due to state.selectedResult being undefined.
      const workflowDagId = state.selectedResult?.workflow_dag_id
        ? state.selectedResult.workflow_dag_id
        : 0;
      state.selectedDag = state.dags[workflowDagId];
    },
  },
  extraReducers: (builder) => {
    builder.addCase(handleGetSelectDagPosition.pending, (state) => {
      state.selectedDagPosition = {
        loadingStatus: { loading: LoadingStatusEnum.Loading, err: '' },
      };
    });
    builder.addCase(handleGetSelectDagPosition.fulfilled, (state, action) => {
      const response = action.payload;
      state.selectedDagPosition.loadingStatus = {
        loading: LoadingStatusEnum.Succeeded,
        err: '',
      };
      state.selectedDagPosition.result = response;
    });
    builder.addCase(handleGetSelectDagPosition.rejected, (state, action) => {
      const payload = action.payload;
      state.selectedDagPosition.loadingStatus = {
        loading: LoadingStatusEnum.Failed,
        err: payload as string,
      };
    });
    builder.addCase(handleGetOperatorResults.pending, (state, action) => {
      const operatorId = action.meta.arg.operatorId;
      state.operatorResults[operatorId] = {
        loadingStatus: { loading: LoadingStatusEnum.Loading, err: '' },
      };
    });
    builder.addCase(handleGetOperatorResults.fulfilled, (state, action) => {
      const operatorId = action.meta.arg.operatorId;
      const response = action.payload;
      state.operatorResults[operatorId].loadingStatus = {
        loading: LoadingStatusEnum.Succeeded,
        err: '',
      };
      state.operatorResults[operatorId].result = response;
    });
    builder.addCase(handleGetOperatorResults.rejected, (state, action) => {
      const operatorId = action.meta.arg.operatorId;
      const payload = action.payload;
      state.operatorResults[operatorId].loadingStatus = {
        loading: LoadingStatusEnum.Failed,
        err: payload as string,
      };
    });

    builder.addCase(handleGetArtifactResults.pending, (state, action) => {
      const artifactId = action.meta.arg.artifactId;
      state.artifactResults[artifactId] = {
        loadingStatus: { loading: LoadingStatusEnum.Loading, err: '' },
      };
    });

    builder.addCase(handleGetArtifactResults.fulfilled, (state, action) => {
      const artifactId = action.meta.arg.artifactId;
      const response = action.payload;
      state.artifactResults[artifactId].loadingStatus = {
        loading: LoadingStatusEnum.Succeeded,
        err: '',
      };
      state.artifactResults[artifactId].result = response;
      if (
        isSucceeded(state.selectedDagPosition.loadingStatus) &&
        !!state.selectedDag
      ) {
        updateNodeWithArtifactResult(
          state.selectedDagPosition.result.nodes,
          state.selectedDag,
          artifactId,
          response.data
        );
      }
    });

    builder.addCase(handleGetArtifactResults.rejected, (state, action) => {
      const artifactId = action.meta.arg.artifactId;
      const payload = action.payload;
      state.artifactResults[artifactId].loadingStatus = {
        loading: LoadingStatusEnum.Failed,
        err: payload as string,
      };
    });

    builder.addCase(handleListWorkflowSavedObjects.pending, (state) => {
      state.savedObjects.loadingStatus = {
        loading: LoadingStatusEnum.Loading,
        err: '',
      };
    });

    builder.addCase(
      handleListWorkflowSavedObjects.fulfilled,
      (state, action) => {
        const response = action.payload;
        const savedObjects = {};

        // Only run this code if there are saved objects. If there are none, then just skip
        // this altogether.
        if (!!response.object_details) {
          response.object_details.map((object: SavedObject) => {
            const key = String([
              object.integration_name,
              getSavedObjectIdentifier(object),
            ]);
            if (savedObjects[key]) {
              savedObjects[key].push(object);
            } else {
              savedObjects[key] = [object];
            }
          });
        }

        state.savedObjects.loadingStatus = {
          loading: LoadingStatusEnum.Succeeded,
          err: '',
        };
        state.savedObjects.result = savedObjects;
      }
    );

    builder.addCase(
      handleListWorkflowSavedObjects.rejected,
      (state, action) => {
        const payload = action.payload;
        state.savedObjects.loadingStatus = {
          loading: LoadingStatusEnum.Failed,
          err: payload as string,
        };
      }
    );

    builder.addCase(handleGetWorkflow.pending, (state) => {
      state.loadingStatus = { loading: LoadingStatusEnum.Loading, err: '' };
    });

    builder.addCase(
      handleGetWorkflow.fulfilled,
      (state, { payload }: PayloadAction<GetWorkflowResponse>) => {
        state.dags = payload.workflow_dags;
        state.dagResults = payload.workflow_dag_results;

        state.artifactResults = {};
        state.operatorResults = {};
        if (!state.selectedResult) {
          state.selectedResult = state.dagResults[0];
        }
        state.selectedDag = state.dags[state.selectedResult.workflow_dag_id];
        state.loadingStatus = { loading: LoadingStatusEnum.Succeeded, err: '' };
      }
    );

    builder.addCase(handleGetWorkflow.rejected, (state, { payload }) => {
      state.loadingStatus = {
        loading: LoadingStatusEnum.Failed,
        err: payload as string,
      };
    });

    builder.addCase(handleDeleteWorkflow.pending, (state) => {
      state.savedObjectDeletion.loadingStatus = {
        loading: LoadingStatusEnum.Loading,
        err: '',
      };
    });

    builder.addCase(handleDeleteWorkflow.rejected, (state, action) => {
      const payload = action.payload;
      state.savedObjectDeletion.loadingStatus = {
        loading: LoadingStatusEnum.Failed,
        err: payload as string,
      };
    });

    builder.addCase(handleDeleteWorkflow.fulfilled, (state, action) => {
      const response = action.payload;
      state.savedObjectDeletion.loadingStatus = {
        loading: LoadingStatusEnum.Succeeded,
        err: '',
      };
      state.savedObjectDeletion.result = response.saved_object_deletion_results;
    });
  },
});

export const { resetState, selectResultIdx } = workflowSlice.actions;
export default workflowSlice.reducer;
