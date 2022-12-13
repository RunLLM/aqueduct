import { faXmark } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Link, ListItem, Typography } from '@mui/material';
import Box from '@mui/material/Box';
import React from 'react';
import { useDispatch } from 'react-redux';

import { handleArchiveNotification } from '../../reducers/notifications';
import { AppDispatch } from '../../stores/store';
import { theme } from '../../styles/theme/theme';
import UserProfile from '../../utils/auth';
import { getPathPrefix } from '../../utils/getPathPrefix';
import { dateString } from '../../utils/metadata';
import { NotificationLogLevel } from '../../utils/notifications';
import { Notification } from '../../utils/notifications';

type Props = {
  user: UserProfile;
  notification: Notification;
};

export const NotificationListItem: React.FC<Props> = ({
  user,
  notification,
}) => {
  const dispatch: AppDispatch = useDispatch();
  const colorMap = {
    [NotificationLogLevel.Success]: 'green.400',
    [NotificationLogLevel.Warning]: 'yellow.400',
    [NotificationLogLevel.Error]: 'red.500',
    [NotificationLogLevel.Info]: 'blue.400',
    [NotificationLogLevel.Neutral]: 'purple.100',
  };

  const borderColor = colorMap[notification.level];

  const content = (
    <Box
      sx={{
        display: 'flex',
        width: '100%',
        alignItems: 'start',
      }}
    >
      <Box sx={{ flex: 1 }}>
        <Link
          underline="none"
          color="inherit"
          href={`${getPathPrefix()}/workflow/${
            notification.workflowMetadata.id
          }?workflowDagResultId=${encodeURI(
            notification.workflowMetadata.dag_result_id
          )}`}
        >
          <Typography
            variant="body1"
            gutterBottom
            sx={{
              '&:hover': { textDecoration: 'underline' },
            }}
          >
            {notification.content}
          </Typography>

          <Typography
            variant="body2"
            sx={{
              fontWeight: 'light',
              color: 'gray.600',
            }}
          >
            {`${dateString(notification.createdAt)}`}
          </Typography>
        </Link>
      </Box>

      <FontAwesomeIcon
        icon={faXmark}
        style={{
          cursor: 'pointer',
          color: 'gray.600',
        }}
        onClick={() =>
          dispatch(handleArchiveNotification({ user, id: notification.id }))
        }
      />
    </Box>
  );
  let notifBackground = theme.palette.TableSuccessBackground;
  if (notification.level === NotificationLogLevel.Warning) {
    notifBackground = theme.palette.TableWarningBackground;
  } else if (notification.level === NotificationLogLevel.Error) {
    notifBackground = theme.palette.TableErrorBackground;
  }
  return (
    <ListItem
      sx={{
        borderLeft: `8px solid`,
        borderColor: borderColor,
        minWidth: '450px',
        maxWidth: '450px',
        '&:hover': {
          backgroundColor: notifBackground,
        },
      }}
    >
      {content}
    </ListItem>
  );
};

export default NotificationListItem;
