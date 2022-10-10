import { faBell, faCircleUser } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
  AppBar,
  Avatar,
  Breadcrumbs,
  Link,
  Menu,
  MenuItem,
  Toolbar,
  Typography,
} from '@mui/material';
import Box from '@mui/material/Box';
import React, { useState } from 'react';
import { useSelector } from 'react-redux';
import { Link as RouterLink } from 'react-router-dom';

import { RootState } from '../../stores/store';
import { theme } from '../../styles/theme/theme';
import UserProfile from '../../utils/auth';
import { getPathPrefix } from '../../utils/getPathPrefix';
import {
  NotificationLogLevel,
  NotificationStatus,
} from '../../utils/notifications';
import NotificationsPopover from '../notifications/NotificationsPopover';
import styles from './menu-sidebar-styles.module.css';
import { MenuSidebarWidthNumber } from './menuSidebar';

const pathPrefix = getPathPrefix();

export class BreadcrumbLinks {
  static readonly HOME = new BreadcrumbLinks(`${pathPrefix}/`, 'Home');
  static readonly DATA = new BreadcrumbLinks(`${pathPrefix}/data`, 'Data');
  static readonly INTEGRATIONS = new BreadcrumbLinks(
    `${pathPrefix}/integrations`,
    'Integrations'
  );
  static readonly WORKFLOWS = new BreadcrumbLinks(
    `${pathPrefix}/workflows`,
    'Workflows'
  );
  static readonly ACCOUNT = new BreadcrumbLinks(
    `${pathPrefix}/account`,
    'Account'
  );
  static readonly ERROR = new BreadcrumbLinks(
    `${pathPrefix}/404`,
    'Page Not Found'
  );

  constructor(public readonly address: string, public readonly name: any) {}

  toString() {
    return this.name;
  }
}

/**
 * The `NavBar` is the core sidebar that we include throughout our UI. It
 * is pinned to the top of every page in our UI, and it includes
 * information about the site hierarchy, notifications, and settings/accounts page.
 */
const NavBar: React.FC<{
  user: UserProfile;
  breadcrumbs: BreadcrumbLinks[];
}> = ({ user, breadcrumbs }) => {
  const [userPopoverAnchorEl, setUserPopoverAnchorEl] = useState(null);
  const [anchorEl, setAnchorEl] = useState(null);

  console.log(breadcrumbs);

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

  const avatarStyling = { width: '24px', height: '24px', marginLeft: '16px' };

  const avatar = user.picture ? (
    <Avatar
      className={styles['user-avatar']}
      sx={avatarStyling}
      src={user.picture}
      onClick={handleUserPopoverClick}
    />
  ) : (
    <Avatar
      className={styles['user-avatar']}
      sx={avatarStyling}
      onClick={handleUserPopoverClick}
    >
      {user.name !== 'aqueduct user' ? user.name : null}
    </Avatar>
  );

  /* Header */
  return (
    <AppBar
      sx={{
        width: `calc(100% - ${MenuSidebarWidthNumber}px)`,
        boxShadow: 'none',
        borderBottom: `2px solid ${theme.palette.gray[300]}`,
        backgroundColor: 'white',
        color: 'black',
      }}
    >
      <Toolbar>
        <Breadcrumbs>
          {breadcrumbs.map((link, index) => {
            if (index + 1 === breadcrumbs.length) {
              return (
                <Typography key={link.name} color="text.primary">
                  {link.name}
                </Typography>
              );
            }
            return (
              <Link
                key={link.name}
                underline="hover"
                color="inherit"
                to={link.address}
                component={RouterLink as any}
              >
                {link.name}
              </Link>
            );
          })}
        </Breadcrumbs>

        <Box sx={{ marginLeft: 'auto' }}>
          <Box onClick={handleClick} sx={{ display: 'flex' }}>
            <Box className={styles['notification-alert']}>
              {!!numUnreadNotifications && (
                <Typography
                  variant="body2"
                  sx={{ fontSize: '12px', fontWeight: 'light', color: 'white' }}
                >
                  {numUnreadNotifications}
                </Typography>
              )}
            </Box>

            <FontAwesomeIcon
              className={styles['menu-sidebar-icon']}
              icon={faBell}
            />
          </Box>

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
              },
            }}
          >
            <Link
              to={`${pathPrefix}/account`}
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
