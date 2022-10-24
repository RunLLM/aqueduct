import { createAsyncThunk, createSlice } from '@reduxjs/toolkit';

import UserProfile from '../utils/auth';
import { archiveNotification, listNotifications } from '../utils/notifications';
import { Notification } from '../utils/notifications';

export interface NotificationsState {
  // fill out notifications state here.
  loading: boolean;
  errorMessage: string;
  notifications: Notification[];
}

const initialNotificationsState: NotificationsState = {
  // fill out initialState
  loading: false,
  errorMessage: '',
  notifications: [],
};

export const handleArchiveNotification = createAsyncThunk<
  // return type of the payload creator
  void,
  // first argument to the payload creator
  { user: UserProfile; id: string },
  // arguments for the ThunkAPI
  {
    rejectValue: string;
  }
>(
  'notificationsReducer/archive',
  async (
    args: {
      user: UserProfile;
      id: string;
    },
    thunkAPI
  ) => {
    const { user, id } = args;
    const errMsg = await archiveNotification(user, id);
    if (errMsg) {
      return thunkAPI.rejectWithValue(errMsg);
    }

    return;
  }
);

export const handleArchiveAllNotifications = createAsyncThunk(
  'notificationsReducer/archiveAll',
  async (args: { user: UserProfile; notifications: Notification[] }) => {
    const { user, notifications } = args;
    await Promise.all(
      notifications.map((notification) =>
        archiveNotification(user, notification.id)
      )
    );

    // We don't handle any error here. In the worst case, user will reload and see some unremoved messages.
    return;
  }
);

export const handleFetchNotifications = createAsyncThunk(
  'notificationsReducer/listNotifications',
  async (args: { user: UserProfile }, thunkAPI) => {
    const { user } = args;

    const [res, errMsg] = await listNotifications(user);
    if (errMsg) {
      return thunkAPI.rejectWithValue(errMsg);
    }
    return res;
  }
);

export const notificationsSlice = createSlice({
  name: 'notificationsReducer',
  initialState: initialNotificationsState,
  reducers: {},
  extraReducers: (builder) => {
    builder.addCase(handleFetchNotifications.pending, (state) => {
      state.loading = true;
    });
    builder.addCase(
      handleFetchNotifications.fulfilled,
      (state, { payload }) => {
        state.notifications = payload as Notification[];
        state.loading = false;
      }
    );
    builder.addCase(handleFetchNotifications.rejected, (state, { payload }) => {
      const errorMessage = payload as string;
      state.errorMessage = errorMessage;
      state.loading = false;
    });
    // archive all notifications
    builder.addCase(handleArchiveAllNotifications.pending, (state) => {
      state.loading = true;
    });
    builder.addCase(handleArchiveAllNotifications.fulfilled, (state) => {
      state.errorMessage = initialNotificationsState.errorMessage;
      state.loading = initialNotificationsState.loading;
      state.notifications = initialNotificationsState.notifications;
    });
    builder.addCase(
      handleArchiveAllNotifications.rejected,
      (state, { payload }) => {
        const errorMessage = payload as string;
        state.errorMessage = errorMessage;
      }
    );
    // archive notification by id.
    builder.addCase(handleArchiveNotification.pending, (state) => {
      state.loading = true;
    });
    builder.addCase(handleArchiveNotification.fulfilled, (state, { meta }) => {
      const notificationId = meta.arg.id;
      state.notifications = state.notifications.filter(
        (notification) => notification.id !== notificationId
      );
      state.loading = false;
    });
    builder.addCase(
      handleArchiveNotification.rejected,
      (state, { payload }) => {
        const errorMessage = payload as string;
        state.loading = false;
        state.errorMessage = errorMessage;
      }
    );
  },
});

export default notificationsSlice.reducer;
