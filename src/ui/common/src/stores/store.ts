import { configureStore } from '@reduxjs/toolkit';

import artifactResultContentsReducer from '../reducers/artifactResultContents';
import artifactResultsReducer from '../reducers/artifactResults';
import dataPreviewReducer from '../reducers/dataPreview';
import integrationReducer from '../reducers/integration';
import integrationsReducer from '../reducers/integrations';
import listWorkflowReducer from '../reducers/listWorkflowSummaries';
import nodeSelectionReducer from '../reducers/nodeSelection';
import notificationsReducer from '../reducers/notifications';
import serverConfigReducer from '../reducers/serverConfig';
import workflowReducer from '../reducers/workflow';
import workflowDagResultsReducer from '../reducers/workflowDagResults';
import workflowDagsReducer from '../reducers/workflowDags';

export const store = configureStore({
  reducer: {
    artifactResultContentsReducer,
    nodeSelectionReducer,
    notificationsReducer,
    listWorkflowReducer,
    dataPreviewReducer,
    integrationsReducer,
    integrationReducer,
    serverConfigReducer,
    workflowReducer,
    workflowDagsReducer,
    workflowDagResultsReducer,
    artifactResultsReducer,
  },
});

// Infer the `RootState` and `AppDispatch` types from the store itself
export type RootState = ReturnType<typeof store.getState>;
// Inferred type: {posts: PostsState, comments: CommentsState, users: UsersState}
export type AppDispatch = typeof store.dispatch;
