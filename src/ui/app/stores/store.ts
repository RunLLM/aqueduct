import dataPreviewReducer from "@reducers/dataPreview";
import integrationsReducer from "@reducers/integrations";
import listWorkflowReducer from "@reducers/listWorkflowSummaries";
import nodeSelectionReducer from "@reducers/nodeSelection";
import notificationsReducer from "@reducers/notifications";
import openSideSheetReducer from "@reducers/openSideSheet";
import workflowReducer from "@reducers/workflow";
import { configureStore } from "@reduxjs/toolkit";

export const store = configureStore({
  reducer: {
    nodeSelectionReducer,
    openSideSheetReducer,
    notificationsReducer,
    listWorkflowReducer,
    dataPreviewReducer,
    integrationsReducer,
    workflowReducer,
  },
});

// Infer the `RootState` and `AppDispatch` types from the store itself
export type RootState = ReturnType<typeof store.getState>;
// Inferred type: {posts: PostsState, comments: CommentsState, users: UsersState}
export type AppDispatch = typeof store.dispatch;
