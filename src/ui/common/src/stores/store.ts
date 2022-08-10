import { configureStore } from '@reduxjs/toolkit';

import dataPreviewReducer from '../reducers/dataPreview';
import integrationReducer from '../reducers/integration';
import integrationsReducer from '../reducers/integrations';
import listWorkflowReducer from '../reducers/listWorkflowSummaries';
import nodeSelectionReducer from '../reducers/nodeSelection';
import notificationsReducer from '../reducers/notifications';
import openSideSheetReducer from '../reducers/openSideSheet';
import workflowReducer from '../reducers/workflow';
//import {AnyAction, CombinedState, configureStore, Reducer} from '@reduxjs/toolkit';

/*
const rootReducer: Reducer<CombinedState<{
    nodeSelectionReducer
    issuesDisplay: CurrentDisplayState;
    repoDetails: RepoDetailsState;
    issues: IssuesState;
    comments: CommentsState;
}>, AnyAction>
*/

export const store = configureStore({
  reducer: {
    nodeSelectionReducer,
    openSideSheetReducer,
    notificationsReducer,
    listWorkflowReducer,
    dataPreviewReducer,
    integrationsReducer,
    integrationReducer,
    workflowReducer,
  },
});

// Infer the `RootState` and `AppDispatch` types from the store itself
export type RootState = ReturnType<typeof store.getState>;
// Inferred type: {posts: PostsState, comments: CommentsState, users: UsersState}
export type AppDispatch = typeof store.dispatch;
