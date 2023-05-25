import { configureStore } from '@reduxjs/toolkit';
import { setupListeners } from '@reduxjs/toolkit/dist/query';

import { aqueductApi } from '../handlers/AqueductApi';
import dataPreviewReducer from '../reducers/dataPreview';
import listWorkflowReducer from '../reducers/listWorkflowSummaries';
import notificationsReducer from '../reducers/notifications';
import workflowPageReducer from '../reducers/pages/Workflow';
import resourceReducer from '../reducers/resource';
import resourcesReducer from '../reducers/resources';
import serverConfigReducer from '../reducers/serverConfig';

export const store = configureStore({
  reducer: {
    [aqueductApi.reducerPath]: aqueductApi.reducer,
    notificationsReducer,
    listWorkflowReducer,
    dataPreviewReducer,
    resourcesReducer,
    resourceReducer,
    serverConfigReducer,
    workflowPageReducer,
  },

  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware().concat(aqueductApi.middleware),
});

setupListeners(store.dispatch);

// Infer the `RootState` and `AppDispatch` types from the store itself
export type RootState = ReturnType<typeof store.getState>;
// Inferred type: {posts: PostsState, comments: CommentsState, users: UsersState}
export type AppDispatch = typeof store.dispatch;
