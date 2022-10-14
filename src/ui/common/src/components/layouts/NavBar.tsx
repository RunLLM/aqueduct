import {
  faBell,
  faCircleUser,
  faGear,
} from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
  AppBar,
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

export class BreadcrumbLink {
  static readonly HOME = new BreadcrumbLink(`${pathPrefix}/`, 'Home');
  static readonly DATA = new BreadcrumbLink(`${pathPrefix}/data`, 'Data');
  static readonly INTEGRATIONS = new BreadcrumbLink(
    `${pathPrefix}/integrations`,
    'Integrations'
  );
  static readonly WORKFLOWS = new BreadcrumbLink(
    `${pathPrefix}/workflows`,
    'Workflows'
  );
  static readonly ACCOUNT = new BreadcrumbLink(
    `${pathPrefix}/account`,
    'Account'
  );
  static readonly ERROR = new BreadcrumbLink(
    `${pathPrefix}/404`,
    'Page Not Found'
  );

  constructor(public readonly address: string, public readonly name: any) { }

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
  breadcrumbs: BreadcrumbLink[];
}> = ({ user, breadcrumbs }) => {
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

  return (
    <AppBar
      sx={{
        width: `calc(100% - ${MenuSidebarWidthNumber}px)`,
        height: '64px',
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
          <Box
            onClick={handleClick}
            sx={{ display: 'flex', cursor: 'pointer' }}
          >
            {!!numUnreadNotifications && (
              <Box className={styles['notification-alert']}>
                <Typography
                  variant="body2"
                  sx={{ fontSize: '12px', fontWeight: 'light', color: 'white' }}
                >
                  {numUnreadNotifications}
                </Typography>
              </Box>
            )}

            <FontAwesomeIcon className={styles['navbar-icon']} icon={faBell} />
          </Box>

          <NotificationsPopover
            user={user}
            id={notificationsPopoverId}
            anchorEl={anchorEl}
            handleClose={handleClose}
            open={open}
          />
        </Box>

        <Box sx={{ cursor: 'pointer', marginLeft: '8px' }}>
          <FontAwesomeIcon
            className={styles['navbar-icon']}
            icon={faGear}
            onClick={handleUserPopoverClick}
          />
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
              sx={{ color: 'blue.900' }}
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
