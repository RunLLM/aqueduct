import { faXmark } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Link, ListItem, Typography } from '@mui/material';
import Box from '@mui/material/Box';

import React from 'react';
import { useDispatch } from 'react-redux';
import {NotificationAssociation, NotificationLogLevel} from "../../utils/notifications";
import UserProfile from "../../utils/auth";
import {handleArchiveNotification} from "../../reducers/notifications";
import {dateString} from "../../utils/metadata";
import {theme} from "../../styles/theme/theme";
import {AppDispatch} from "../../stores/store";
import {Notification} from '../../utils/notifications';

type Props = {
    user: UserProfile;
    notification: Notification;
};

export const NotificationListItem: React.FC<Props> = ({ user, notification }) => {
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
                    <Typography variant="body1" sx={{ color: 'gray800' }}>
                        <Typography
                            variant="body1"
                            gutterBottom
                            sx={{ fontFamily: 'Monospace', '&:hover': { textDecoration: 'underline' } }}
                        >
                            {notification.workflowMetadata.name}
                        </Typography>

                        {notification.level === NotificationLogLevel.Success ? ' succeeded' : ' failed'}
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
                {title}
                <FontAwesomeIcon
                    icon={faXmark}
                    style={{
                        cursor: 'pointer',
                        color: 'gray.600',
                    }}
                    onClick={() => dispatch(handleArchiveNotification({ user, id: notification.id }))}
                />
            </Box>

            <Typography
                variant="h6"
                gutterBottom
                sx={{ fontFamily: 'Monospace', '&:hover': { textDecoration: 'underline' } }}
            >
                {notification.content}
            </Typography>

            <Box
                sx={{
                    display: 'flex',
                    flexDirection: 'row',
                    justifyContent: 'space-between',
                }}
            >
                <Box>
                    <Typography variant="body1" sx={{ fontWeight: 'medium', color: 'gray600' }}>
                        {/* Show notification associated with 'workflow_run' as 'workflow' category */}
                        {association.object === NotificationAssociation.WorkflowDagResult
                            ? NotificationAssociation.Workflow
                            : association.object}
                    </Typography>
                </Box>
                <Box>
                    <Typography variant="body1" sx={{ fontWeight: 'light', color: 'gray600' }}>
                        {`${dateString(notification.createdAt)}`}
                    </Typography>
                </Box>
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
            href={`/workflow/${notification.workflowMetadata.id}/?workflowDagResultId=${encodeURI(
                notification.workflowMetadata.dag_result_id,
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
