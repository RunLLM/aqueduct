import {
  faBell,
  faCircleUser,
  faDatabase,
  faMessage,
  faPlug,
  faShareNodes,
} from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { AppBar, Avatar, Breadcrumbs, Link, Menu, MenuItem, Toolbar, Typography } from '@mui/material';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Divider from '@mui/material/Divider';
import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Link as RouterLink, useLocation } from 'react-router-dom';
import { theme } from '../../styles/theme/theme';

import { handleFetchNotifications } from '../../reducers/notifications';
import { AppDispatch, RootState } from '../../stores/store';
import UserProfile from '../../utils/auth';
import { getPathPrefix } from '../../utils/getPathPrefix';
import {
  NotificationLogLevel,
  NotificationStatus,
} from '../../utils/notifications';
import NotificationsPopover from '../notifications/NotificationsPopover';
import styles from './menu-sidebar-styles.module.css';
import { MenuSidebarWidthNumber } from './menuSidebar';

const BUTTON_STYLE_OVERRIDE = {
  display: 'flex',
  flexDirection: 'row',
  alignItems: 'center',
  cursor: 'pointer',
  justifyContent: 'left',
  paddingX: 1,
  width: '100%',
  maxWidth: '100%',
  textTransform: 'none',
} as const;

/**
 * The `NavBar` is the core sidebar that we include throughout our UI. It
 * is pinned to the top of every page in our UI, and it includes
 * information about the site hierarchy, notifications, and settings/accounts page.
 */
const NavBar: React.FC<{ user: UserProfile }> = ({ user }) => {
  const [userPopoverAnchorEl, setUserPopoverAnchorEl] = useState(null);
  const [anchorEl, setAnchorEl] = useState(null);
  
  const userPopoverOpen = Boolean(userPopoverAnchorEl);
  const open = Boolean(anchorEl);

  const numUnreadNotifications = useSelector(
    (state: RootState) =>
      state.notificationsReducer.notifications.filter(
        (notification) =>
          notification.level !== NotificationLogLevel.Success &&
          notification.status === NotificationStatus.Unread
      ).length
  );

  const handleUserPopoverClick = (event: React.MouseEvent) => {
    setUserPopoverAnchorEl(event.currentTarget);
  };

  const handleCloseUserPopover = () => {
    setUserPopoverAnchorEl(null);
  };

  const handleClick = (event: React.MouseEvent<HTMLDivElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };
  
  const notificationsPopoverId = open ? 'simple-popover' : undefined;
  const userPopoverId = userPopoverOpen ? 'user-popover' : undefined;

  const avatar = user.picture ? (
    <Avatar
      className={styles['user-avatar']}
      sx={{ width: '24px', height: '24px' }}
      src={user.picture}
      onClick={handleUserPopoverClick}
    />
  ) : (
    <Avatar
      className={styles['user-avatar']}
      sx={{ width: '24px', height: '24px' }}
      onClick={handleUserPopoverClick}
    >
      {user.name !== 'aqueduct user' ? user.name : null}
    </Avatar>
  );

  /* Header */
  return (
    <AppBar sx={{
      width: `calc(100% - ${MenuSidebarWidthNumber}px)`,
      boxShadow: 'none',
      borderBottom: `2px solid ${theme.palette.gray[300]}`,
      backgroundColor: 'white',
      color: 'black'
      }}>
      <Toolbar>
      <Breadcrumbs>
        <Link
          underline="hover"
          color="inherit"
          to="/"
          component={RouterLink as any}
        >
          Home
        </Link>
        <Typography color="text.primary">Integrations</Typography>
      </Breadcrumbs>

      <Box onClick={handleClick} sx={{display: 'flex', marginLeft: 'auto'}}>
        <Box className={styles['notification-alert']}>
              { !!numUnreadNotifications && <Typography
                variant="body2"
                sx={{ fontSize: '12px', fontWeight: 'light', color: 'white' }}
              >
                {numUnreadNotifications}
              </Typography>}
        </Box>

        <FontAwesomeIcon
          className={styles['menu-sidebar-icon']}
          icon={faBell}
        />

        <NotificationsPopover
          user={user}
          id={notificationsPopoverId}
          anchorEl={anchorEl}
          handleClose={handleClose}
          open={open}
        />
      </Box>

      <Box>
          {avatar}
          <Menu
            id={userPopoverId}
            anchorEl={userPopoverAnchorEl}
            onClose={handleCloseUserPopover}
            open={userPopoverOpen}
            PaperProps={{
              sx: {
                mt: 1.5,
              }
            }}
          >
            <Link
              to={`${getPathPrefix()}/account`}
              underline="none"
              sx={{ color: 'blue.800' }}
              component={RouterLink as any}
            >
              <MenuItem sx={{ width: '190px' }} disableRipple>
                <Box sx={{ fontSize: '20px', mr: 1 }}>
                  <FontAwesomeIcon icon={faCircleUser} />
                </Box>
                Account
              </MenuItem>
            </Link>
          </Menu>
      </Box>
      </Toolbar>
    </AppBar>
  );
};

export default NavBar;
