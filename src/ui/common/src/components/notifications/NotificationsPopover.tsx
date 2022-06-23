import { faBoxArchive, faInbox } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { List, Popover, Typography } from '@mui/material';
import Box from '@mui/material/Box';
import React, { useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { handleArchiveAllNotifications } from '../../reducers/notifications';
import { AppDispatch, RootState } from '../../stores/store';
import UserProfile from '../../utils/auth';
import { NotificationAssociation, NotificationLogLevel} from '../../utils/notifications';
import { Notification } from '../../utils/notifications';
import { Tab, Tabs } from '../primitives/Tabs.styles';
import NotificationListItem from './NotificationListItem';

interface NotificationsPopoverProps {
  user: UserProfile;
  id: string;
  anchorEl: Element | null;
  handleClose: () => void;
  open: boolean;
}

enum NotificationsTabs {
  All = 'All',
  Workflow = 'Workflow',
  Team = 'Team',
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
  const [activeTab, setActiveTab] = useState(NotificationsTabs.All);
  const notifications: Notification[] = notificationsState.notifications;

  const handleChangeTab = (event: React.SyntheticEvent, newValue: string) => {
    event.preventDefault();
    setActiveTab(NotificationsTabs[newValue]);
  };

  const handleArchiveAll = (event: React.MouseEvent) => {
    event.preventDefault();
    dispatch(handleArchiveAllNotifications({ user, notifications }));
  };

  const filteredNotifications: Notification[] = notifications
    .filter((notification: Notification) => {
      if (notification.level === NotificationLogLevel.Success) {
        return false
      }
      const association = notification.association.object;
      if (activeTab === NotificationsTabs.All) {
        return true;
      } else if (activeTab === NotificationsTabs.Workflow) {
        return (
          association === NotificationAssociation.Workflow ||
          association === NotificationAssociation.WorkflowDagResult
        );
      } else if (activeTab === NotificationsTabs.Team) {
        return association === NotificationAssociation.Organization;
      }

      return true;
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
    >
      <Tabs value={activeTab} onChange={handleChangeTab}>
        <Tab
          key={NotificationsTabs.All}
          label={NotificationsTabs.All}
          value={NotificationsTabs.All}
          sx={{ '&:hover': { color: 'gray900' } }}
        />
        <Tab
          key={NotificationsTabs.Workflow}
          label={NotificationsTabs.Workflow}
          value={NotificationsTabs.Workflow}
          sx={{ '&:hover': { color: 'gray900' } }}
        />
        <Tab
          key={NotificationsTabs.Team}
          label={NotificationsTabs.Team}
          value={NotificationsTabs.Team}
          sx={{ '&:hover': { color: 'gray900' } }}
        />
      </Tabs>

      <Box role="tabpanel" sx={{ minHeight: '400px' }}>
        {filteredNotifications.length > 0 ? (
          <List>
            {filteredNotifications.map((notification, _) => {
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
          width: '100%',
          height: '36px',
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
