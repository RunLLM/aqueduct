//import { dataPreview as dataPreviewReducer } from '@aqueducthq/common';
import { integrations as integrationsReducer } from '@aqueducthq/common';
import { workflowSummaries as listWorkflowReducer } from '@aqueducthq/common';
import { integration as integrationReducer } from '@aqueducthq/common';
import { nodeSelection as nodeSelectionReducer } from '@aqueducthq/common';
import { notifications as notificationsReducer } from '@aqueducthq/common';
import { openSideSheet as openSideSheetReducer } from '@aqueducthq/common';
import { workflow as workflowReducer } from '@aqueducthq/common';
import { workflowDagResults as workflowDagResultsReducer } from '@aqueducthq/common';
import { artifactResultContents as artifactResultContentsReducer } from '@aqueducthq/common';
import { artifactResults as artifactResultsReducer } from '@aqueducthq/common';
import { configureStore } from '@reduxjs/toolkit';

export const store = configureStore({
    reducer: {
        nodeSelectionReducer,
        openSideSheetReducer,
        notificationsReducer,
        listWorkflowReducer,
        //dataPreviewReducer,
        integrationReducer,
        integrationsReducer,
        workflowReducer,
        workflowDagResultsReducer,
        artifactResultsReducer,
        artifactResultContentsReducer,
    },
});

// Infer the `RootState` and `AppDispatch` types from the store itself
export type RootState = ReturnType<typeof store.getState>;
// Inferred type: {posts: PostsState, comments: CommentsState, users: UsersState}
export type AppDispatch = typeof store.dispatch;
