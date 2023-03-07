// This is being deprecated. Please use `workflowDagResults` combining with
// `artifactResults` for future development.
import { createAsyncThunk, createSlice, PayloadAction } from '@reduxjs/toolkit';
import ELK from 'elkjs/lib/elk.bundled.js';
import { Edge, Node } from 'react-flow-renderer';

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
  {
    apiKey: string;
    workflowDagResultId: string;
    artifactId: string;
    metadataOnly: boolean;
  }
>(
  'workflowReducer/getArtifactResults',
  async (
    args: {
      apiKey: string;
      workflowDagResultId: string;
      artifactId: string;
      metadataOnly: boolean;
    },
    thunkAPI
  ) => {
    const { apiKey, workflowDagResultId, artifactId, metadataOnly } = args;
    const res = await fetch(
      `${apiAddress}/api/artifact/${workflowDagResultId}/${artifactId}/result`,
      {
        method: 'GET',
        headers: {
          'api-key': apiKey,
          'metadata-only': metadataOnly.toString(),
        },
      }
    );

    try {
      const formData = await res.formData();
      const metadataJson = await (formData.get('metadata') as File).text();
      const artifactResult = JSON.parse(
        metadataJson
      ) as GetArtifactResultResponse;

      if (metadataOnly) {
        return artifactResult;
      }

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
    const { operators, artifacts, onChange, onConnect } = args;
    const opNodes = Object.values(operators)
      .filter((op) => {
        return op.spec.type != OperatorType.Param;
      })
      .map((op) => getOperatorNode(op, onChange, onConnect));
    const artfNodes = Object.values(artifacts).map((artf) =>
      getArtifactNode(artf, onChange, onConnect)
    );
    const edges = getEdges(operators);
    const allNodes = {
      nodes: opNodes.concat(artfNodes),
      edges: edges,
    } as positionResponse;
    const dag = thunkAPI.getState().workflowReducer.selectedDag;
    const artifactResults =
      thunkAPI.getState().workflowReducer.artifactResults ?? {};

    if (!!dag) {
      const collapsedPosition = collapsePosition(
        allNodes,
        dag,
        artifactResults
      );

      const elk = new ELK();
      const mappedNodes = collapsedPosition.nodes.map((node) => {
        return {
          id: node.id,
          // width of the node in px.
          width: 310,
          // height of the node in px.
          height: 120,
        };
      });

      const graph = {
        id: 'root',
        layoutOptions: {
          'elk.algorithm': 'layered',
          'elk.direction': 'RIGHT',
          'elk.spacing.nodeNode': '200',
          'elk.layered.spacing.nodeNodeBetweenLayers': '200',
          'elk.layered.nodePlacement.strategy': 'INTERACTIVE',
        },
        children: mappedNodes,
        edges: collapsedPosition.edges,
      };

      try {
        const positionedLayout = await elk.layout(graph);
        collapsedPosition.nodes = collapsedPosition.nodes.map((node) => {
          for (let i = 0; i < positionedLayout.children.length; i++) {
            if (node.id === positionedLayout.children[i].id) {
              node.position = {
                x: positionedLayout.children[i].x,
                y: positionedLayout.children[i].y,
              };

              // TODO (ENG-2305): Handle mapping elk edge sections to react-flow edges
            }
          }

          return node;
        });
      } catch (error) {
        // TODO (ENG-2304): Show alert box when positioned layout fails to be retrieved.
        console.log('error getting PositionedLayout: ', error);
      }

      return collapsedPosition;
    }

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
