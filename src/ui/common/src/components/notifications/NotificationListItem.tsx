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
import { dateString } from '../../utils/metadata';
import {
  NotificationAssociation,
  NotificationLogLevel,
} from '../../utils/notifications';
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
  const association = notification.association;

  const colorMap = {
    [NotificationLogLevel.Success]: 'green.400',
    [NotificationLogLevel.Warning]: 'yellow.400',
    [NotificationLogLevel.Error]: 'red.500',
    [NotificationLogLevel.Info]: 'blue.400',
    [NotificationLogLevel.Neutral]: 'purple.100',
  };

  const borderColor = colorMap[notification.level];

  let title;
  switch (association.object) {
    case NotificationAssociation.Workflow:
    case NotificationAssociation.WorkflowDagResult: {
      title = !!notification.workflowMetadata ? (
        <Box
          sx={{
            display: 'flex',
            flexDirection: 'row',
          }}
        >
          <Typography variant="body1" sx={{ color: 'gray.800' }}>
            <Typography
              variant="body1"
              gutterBottom
              sx={{
                fontFamily: 'Monospace',
                '&:hover': { textDecoration: 'underline' },
              }}
            >
              {notification.workflowMetadata.name}
            </Typography>
          </Typography>
        </Box>
      ) : (
        <Box />
      );
      break;
    }

    default:
      title = <Box />;
  }

  const content = (
    <Box
      sx={{
        display: 'flex',
        flexDirection: 'column',
        width: '100%',
      }}
    >
      <Box
        sx={{
          display: 'flex',
          flexDirection: 'row',
          justifyContent: 'space-between',
          width: '100%',
        }}
      >
        <Typography
          variant="h6"
          gutterBottom
          sx={{
            fontFamily: 'Monospace',
            '&:hover': { textDecoration: 'underline' },
          }}
        >
          {notification.content}
        </Typography>
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

      <Box
        sx={{
          display: 'flex',
          justifyContent: 'flex-end',
        }}
      >
        <Typography
          variant="body1"
          sx={{
            fontWeight: 'light',
            color: 'gray.600',
          }}
        >
          {`${dateString(notification.createdAt)}`}
        </Typography>
      </Box>
    </Box>
  );
  let notifBackground = theme.palette.TableSuccessBackground;
  if (notification.level === NotificationLogLevel.Warning) {
    notifBackground = theme.palette.TableWarningBackground;
  } else if (notification.level === NotificationLogLevel.Error) {
    notifBackground = theme.palette.TableErrorBackground;
  }
  return (
    <Link
      underline="none"
      color="inherit"
      href={`/workflow/${notification.workflowMetadata.id
        }/?workflowDagResultId=${encodeURI(
          notification.workflowMetadata.dag_result_id
        )}`}
    >
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
    </Link>
  );
};

export default NotificationListItem;
