import { configureStore } from '@reduxjs/toolkit';

import artifactResultContentsReducer from '../reducers/artifactResultContents';
import artifactResultsReducer from '../reducers/artifactResults';
//import dataPreviewReducer from '../reducers/dataPreview';
import integrationReducer from '../reducers/integration';
import integrationsReducer from '../reducers/integrations';
import listWorkflowReducer from '../reducers/listWorkflowSummaries';
import nodeSelectionReducer from '../reducers/nodeSelection';
import notificationsReducer from '../reducers/notifications';
import openSideSheetReducer from '../reducers/openSideSheet';
import workflowReducer from '../reducers/workflow';
import workflowDagResultsReducer from '../reducers/workflowDagResults';

export const store = configureStore({
  reducer: {
    artifactResultContentsReducer,
    nodeSelectionReducer,
    openSideSheetReducer,
    notificationsReducer,
    listWorkflowReducer,
    //dataPreviewReducer,
    integrationsReducer,
    integrationReducer,
    workflowReducer,
    workflowDagResultsReducer,
    artifactResultsReducer,
  },
});

// Infer the `RootState` and `AppDispatch` types from the store itself
export type RootState = ReturnType<typeof store.getState>;
// Inferred type: {posts: PostsState, comments: CommentsState, users: UsersState}
export type AppDispatch = typeof store.dispatch;
