import { aqueductApi } from '@aqueducthq/common';
import { dataPreview as dataPreviewReducer } from '@aqueducthq/common';
import { resource as resourceReducer } from '@aqueducthq/common';
import { resources as resourcesReducer } from '@aqueducthq/common';
import { workflowSummaries as listWorkflowReducer } from '@aqueducthq/common';
import { notifications as notificationsReducer } from '@aqueducthq/common';
import { serverConfig as serverConfigReducer } from '@aqueducthq/common';
import { workflowPage as workflowPageReducer } from '@aqueducthq/common';
import { configureStore } from '@reduxjs/toolkit';
import { setupListeners } from '@reduxjs/toolkit/dist/query';

export const store = configureStore({
    reducer: {
        [aqueductApi.reducerPath]: aqueductApi.reducer,
        notificationsReducer,
        listWorkflowReducer,
        dataPreviewReducer,
        resourceReducer,
        resourcesReducer,
        serverConfigReducer,
        workflowPageReducer,
    },

    middleware: (getDefaultMiddleware) => getDefaultMiddleware().concat(aqueductApi.middleware),
});

setupListeners(store.dispatch);

// Infer the `RootState` and `AppDispatch` types from the store itself
export type RootState = ReturnType<typeof store.getState>;
// Inferred type: {posts: PostsState, comments: CommentsState, users: UsersState}
export type AppDispatch = typeof store.dispatch;
