import { createAsyncThunk, createSlice, PayloadAction } from '@reduxjs/toolkit';

import { useAqueductConsts } from '../components/hooks/useAqueductConsts';
import { GetArtifactResultResponse } from '../utils/artifacts';
import { GetOperatorResultResponse, Operator } from '../utils/operators';
import { LoadingStatus, LoadingStatusEnum } from '../utils/shared';
import { GetPositionResponse } from '../utils/reactflow';
import {
  GetWorkflowResponse,
  normalizeGetWorkflowResponse,
  WorkflowDag,
  WorkflowDagResultSummary,
} from '../utils/workflows';

const { apiAddress } = useAqueductConsts();

type selectDagPositionResult = {
  loadingStatus: LoadingStatus;
  result?: GetPositionResponse
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
  selectedDagPosition? : selectDagPositionResult;
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

export const handleGetSelectDagPosition = createAsyncThunk<
  GetPositionResponse,
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
    console.log("reached handle Position");
    const { apiKey, operators } = args;
    const res = await fetch(
      `${apiAddress}/api/positioning`,
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
    return body as GetPositionResponse;
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

    builder.addCase(handleGetWorkflow.pending, (state) => {
      state.loadingStatus = { loading: LoadingStatusEnum.Loading, err: '' };
    });

    builder.addCase(
      handleGetWorkflow.fulfilled,
      (state, { payload }: PayloadAction<GetWorkflowResponse>) => {
        state.dags = payload.workflow_dags;
        state.dagResults = payload.workflow_dag_results;
        state.watcherAuthIds = payload.watcherAuthIds;

        handleSelectResultIdx(state, 0);
        state.loadingStatus = { loading: LoadingStatusEnum.Succeeded, err: '' };
      }
    );

    builder.addCase(handleGetWorkflow.rejected, (state, { payload }) => {
      state.loadingStatus = {
        loading: LoadingStatusEnum.Failed,
        err: payload as string,
      };
    });
  },
});

export const { selectResultIdx } = workflowSlice.actions;
export default workflowSlice.reducer;
