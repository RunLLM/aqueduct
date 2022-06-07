import { createAsyncThunk, createSlice, PayloadAction } from '@reduxjs/toolkit';

import { useAqueductConsts } from '../components/hooks/useAqueductConsts';
import { GetArtifactResultResponse } from '../utils/artifacts';
import { GetOperatorResultResponse, Operator } from '../utils/operators';
import { LoadingStatus, LoadingStatusEnum } from '../utils/shared';
import { NodePos } from '../utils/reactflow';
import {
  GetWorkflowResponse,
  normalizeGetWorkflowResponse,
  WorkflowDag,
  WorkflowDagResultSummary,
} from '../utils/workflows';

const { httpProtocol, apiAddress } = useAqueductConsts();


type positionResponse = {
  operator_positions: { [opId: string]: NodePos }
  artifact_positions: { [artfId: string]: NodePos }
}

type selectDagPositionResult = {
  loadingStatus: LoadingStatus;
  // result?: [{ [opId: string]: NodePos }, { [artfId: string]: NodePos }]
  result?: positionResponse
}

export type ArtifactResult = {
  loadingStatus: LoadingStatus;
  result?: GetArtifactResultResponse;
};

export type OperatorResult = {
  loadingStatus: LoadingStatus;
  result?: GetOperatorResultResponse;
};

export type WorkflowState = {
  loadingStatus: LoadingStatus;
  dags: { [id: string]: WorkflowDag };
  dagResults: WorkflowDagResultSummary[];
  watcherAuthIds: string[];

  selectedResult?: WorkflowDagResultSummary;
  selectedDag?: WorkflowDag;
  selectedDagPosition?: any;
  artifactResults: { [id: string]: ArtifactResult };
  operatorResults: { [id: string]: OperatorResult };
};


const initialState: WorkflowState = {
  loadingStatus: { loading: LoadingStatusEnum.Initial, err: '' },
  dags: {},
  dagResults: [],
  artifactResults: {},
  operatorResults: {},
  watcherAuthIds: [],
  selectedDagPosition: { loadingStatus: { loading: LoadingStatusEnum.Initial, err: '' }},
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
      `${httpProtocol}://${apiAddress}/operator_result/${workflowDagResultId}/${operatorId}`,
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
      `${httpProtocol}://${apiAddress}/artifact_result/${workflowDagResultId}/${artifactId}`,
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

    const res = await fetch(
      `${httpProtocol}://${apiAddress}/workflow/${workflowId}`,
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

    return normalizeGetWorkflowResponse(body);
  }
);

export const handleGetSelectDagPosition = createAsyncThunk<
  positionResponse,
  { apiKey: string; operators: { [id: string]: Operator } }
>(
  'workflowReducer/getSelectDagPosition',
  async (
    args: {
      apiKey: string;
      operators: { [id: string]: Operator };
    },
    thunkAPI
  ) => {
    const { apiKey, operators } = args;
    const res = await fetch(
      `${httpProtocol}://${apiAddress}/positioning`,
      {
        method: 'POST',
        headers: {
          'api-key': apiKey,
        },
        body: JSON.stringify(operators),
      }
    );

    const body = await res.json();
    if (!res.ok) {
      return thunkAPI.rejectWithValue(body.error);
    }
    console.log("body: ", body)

    return body as positionResponse;
  }
);

const handleSelectResultIdx = (state: WorkflowState, idx: number) => {
  state.artifactResults = {};
  state.operatorResults = {};
  state.selectedResult = state.dagResults[idx];
  state.selectedDag = state.dags[state.selectedResult.workflow_dag_id];
};

export const workflowSlice = createSlice({
  name: 'workflowReducer',
  initialState,
  reducers: {
    selectResultIdx: (state, { payload }: PayloadAction<number>) => {
      handleSelectResultIdx(state, payload);
    },
  },
  extraReducers: (builder) => {
    builder.addCase(handleGetSelectDagPosition.pending, (state, action) => {
      console.log("reached pending")
      state.selectedDagPosition = {
        loadingStatus: { loading: LoadingStatusEnum.Loading, err: '' },
      };
    });
    builder.addCase(handleGetSelectDagPosition.fulfilled, (state, action) => {
      console.log("reached fulfilled")
      const response = action.payload;
      state.selectedDagPosition.loadingStatus = {
        loading: LoadingStatusEnum.Succeeded,
        err: '',
      };
      state.selectedDagPosition.result = response;
    });
    builder.addCase(handleGetSelectDagPosition.rejected, (state, action) => {
      console.log("reached rejected")
      const payload = action.payload;
      state.selectedDagPosition.loadingStatus = {
        loading: LoadingStatusEnum.Failed,
        err: payload as string,
      };
    });
    builder.addCase(handleGetOperatorResults.pending, (state, action) => {
      console.log("reached fulfilled")
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

    builder.addCase(handleGetWorkflow.pending, (state) => {
      console.log("reached fulfilled")
      state.loadingStatus = { loading: LoadingStatusEnum.Loading, err: '' };
    });

    builder.addCase(
      handleGetWorkflow.fulfilled,
      (state, { payload }: PayloadAction<GetWorkflowResponse>) => {
        console.log("reached fulfilled")
        state.dags = payload.workflow_dags;
        state.dagResults = payload.workflow_dag_results;
        state.watcherAuthIds = payload.watcherAuthIds;

        handleSelectResultIdx(state, 0);
        state.loadingStatus = { loading: LoadingStatusEnum.Succeeded, err: '' };
      }
    );

    builder.addCase(handleGetWorkflow.rejected, (state, { payload }) => {
      console.log("reached fulfilled")
      state.loadingStatus = {
        loading: LoadingStatusEnum.Failed,
        err: payload as string,
      };
    });
  },
});

export const { selectResultIdx } = workflowSlice.actions;
export default workflowSlice.reducer;
