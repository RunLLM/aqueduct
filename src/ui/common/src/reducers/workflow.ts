import { createAsyncThunk, createSlice, PayloadAction } from '@reduxjs/toolkit';
import { Edge, Node } from 'react-flow-renderer';

import { useAqueductConsts } from '../components/hooks/useAqueductConsts';
import { Artifact, GetArtifactResultResponse } from '../utils/artifacts';
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
import { LoadingStatus, LoadingStatusEnum } from '../utils/shared';
import {
  DeleteWorkflowResponse,
  GetWorkflowResponse,
  ListWorkflowSavedObjectsResponse,
  normalizeGetWorkflowResponse,
  SavedObject,
  SavedObjectDeletion,
  WorkflowDag,
  WorkflowDagResultSummary,
} from '../utils/workflows';

const { apiAddress } = useAqueductConsts();

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
  result:  Record<string, SavedObjectDeletion[]>;
};


export type WorkflowState = {
  loadingStatus: LoadingStatus;
  savedObjects: SavedObjectResult;
  savedObjectDeletion: SavedObjectDeletionResult;
  dags: { [id: string]: WorkflowDag };
  dagResults: WorkflowDagResultSummary[];
  watcherAuthIds: string[];

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
  watcherAuthIds: [],
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
      `${apiAddress}/api/operator_result/${workflowDagResultId}/${operatorId}`,
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
      `${apiAddress}/api/artifact_result/${workflowDagResultId}/${artifactId}`,
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

    return body as GetArtifactResultResponse;
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
          object.object_name
        );
      } else {
        data['external_delete'][object.integration_name] = [object.object_name];
      }
    });

    const res = await fetch(`${apiAddress}/api/workflow/${workflowId}/delete`, {
          method: 'POST',
          headers: {
            'api-key': apiKey,
          },
          body: JSON.stringify(data),
        })

    const body = await res.json();
    if (!res.ok) {
      return thunkAPI.rejectWithValue(body.error);
    }

    return body as DeleteWorkflowResponse;
  }
);

export const handleGetSelectDagPosition = createAsyncThunk<
  positionResponse,
  {
    apiKey: string;
    operators: { [id: string]: Operator };
    artifacts: { [id: string]: Artifact };
  }
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
    return {
      nodes: opNodes.concat(artfNodes),
      edges: edges,
    } as positionResponse;
  }
);

export const workflowSlice = createSlice({
  name: 'workflowReducer',
  initialState,
  reducers: {
    selectResultIdx: (state, { payload }: PayloadAction<number>) => {
      state.artifactResults = {};
      state.operatorResults = {};
      state.selectedResult = state.dagResults[payload];
      state.selectedDag = state.dags[state.selectedResult.workflow_dag_id];
    },
  },
  extraReducers: (builder) => {
    builder.addCase(handleGetSelectDagPosition.pending, (state, action) => {
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
    });

    builder.addCase(handleGetArtifactResults.rejected, (state, action) => {
      const artifactId = action.meta.arg.artifactId;
      const payload = action.payload;
      state.artifactResults[artifactId].loadingStatus = {
        loading: LoadingStatusEnum.Failed,
        err: payload as string,
      };
    });

    builder.addCase(
      handleListWorkflowSavedObjects.fulfilled,
      (state, action) => {
        const response = action.payload;
        const savedObjects = {};
        response.object_details.map((object) => {
          const key = String([object.integration_name, object.object_name]);
          if (savedObjects[key]) {
            savedObjects[key].push(object);
          } else {
            savedObjects[key] = [object];
          }
        });

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
        state.watcherAuthIds = payload.watcherAuthIds;

        state.artifactResults = {};
        state.operatorResults = {};
        state.selectedResult = state.dagResults[0];
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

export const { selectResultIdx } = workflowSlice.actions;
export default workflowSlice.reducer;
