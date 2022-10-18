import { faBoxArchive, faInbox } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { List, Popover, Typography } from '@mui/material';
import Box from '@mui/material/Box';
import React from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { handleArchiveAllNotifications } from '../../reducers/notifications';
import { AppDispatch, RootState } from '../../stores/store';
import UserProfile from '../../utils/auth';
import { NotificationLogLevel } from '../../utils/notifications';
import { Notification } from '../../utils/notifications';
import NotificationListItem from './NotificationListItem';

export const breadcrumbsSize = '64px';

interface NotificationsPopoverProps {
  user: UserProfile;
  id: string;
  anchorEl: Element | null;
  handleClose: () => void;
  open: boolean;
}

export const NotificationsPopover: React.FC<NotificationsPopoverProps> = ({
  user,
  id,
  anchorEl,
  handleClose,
  open,
}) => {
  const dispatch: AppDispatch = useDispatch();
  const notificationsState = useSelector(
    (state: RootState) => state.notificationsReducer
  );
  const notifications: Notification[] = notificationsState.notifications;

  const handleArchiveAll = (event: React.MouseEvent) => {
    event.preventDefault();
    dispatch(handleArchiveAllNotifications({ user, notifications }));
  };

  const filteredNotifications: Notification[] = notifications
    .filter((notification: Notification) => {
      return notification.level !== NotificationLogLevel.Success;
    })
    .sort((a: Notification, b: Notification) => {
      // Sort notifications in reverse chronological order so that we render most recent notifications first.
      return b.createdAt - a.createdAt;
    });

  return (
    <Popover
      id={id}
      open={open}
      anchorEl={anchorEl}
      onClose={handleClose}
      anchorOrigin={{
        vertical: 'top',
        horizontal: 'right',
      }}
      sx={{ maxHeight: `calc(100% - ${breadcrumbsSize})` }}
      PaperProps={{
        sx: {
          mt: 4.5,
        },
      }}
    >
      <Box role="tabpanel" sx={{ minHeight: '400px' }}>
        {filteredNotifications.length > 0 ? (
          <List>
            {filteredNotifications.map((notification) => {
              return (
                <NotificationListItem
                  user={user}
                  key={notification.id}
                  notification={notification}
                />
              );
            })}
          </List>
        ) : (
          <Box
            sx={{
              display: 'flex',
              justifyContent: 'center',
              flexDirection: 'column',
              alignItems: 'center',
              height: '400px',
              minWidth: '450px',
              color: 'gray.700',
            }}
          >
            <FontAwesomeIcon icon={faInbox} style={{ fontSize: '40px' }} />
            <Typography variant="h5" sx={{ marginTop: '8px' }}>
              Nothing New!
            </Typography>
            <Typography variant="body1" sx={{ marginTop: '8px' }}>
              No notifications at this time.
            </Typography>
          </Box>
        )}
      </Box>

      <Box
        sx={{
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          borderTop: `1px solid`,
          borderColor: 'gray.700',
          position: 'sticky',
          bottom: '0',
          width: '100%',
          height: '36px',
          backgroundColor: 'white',
          opacity: '1',
          color: 'gray.700',
          '&:hover': { backgroundColor: 'blueTint' },
          '&:active': { backgroundColor: 'blue.100' },
        }}
        onClick={handleArchiveAll}
      >
        <FontAwesomeIcon icon={faBoxArchive} />
        <Typography sx={{ ml: 1 }} variant="body1">
          Archive All
        </Typography>
      </Box>
    </Popover>
  );
};

export default NotificationsPopover;
