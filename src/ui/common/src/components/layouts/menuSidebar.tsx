import {
  faBell,
  faCircleUser,
  faDatabase,
  faMessage,
  faPlug,
  faShareNodes,
} from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Avatar, Link, Menu, MenuItem, Typography } from '@mui/material';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Divider from '@mui/material/Divider';
import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Link as RouterLink, useLocation } from 'react-router-dom';

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

// Left padding = 13px
// Right padding = 13px
// Content size = 50px
export const MenuSidebarWidthNumber = 76;
export const MenuSidebarWidth = `${MenuSidebarWidthNumber}px`;

export type SidebarButtonProps = {
  icon: React.ReactElement;
  text: string;
  selected?: boolean;
  numUpdates?: number;
};

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

const SidebarButton: React.FC<SidebarButtonProps> = ({
  icon,
  text,
  numUpdates = 0,
  selected = false,
}) => {
  return (
    <Button
      sx={{
        my: 1,
        ...BUTTON_STYLE_OVERRIDE,
        bg: 'blue.800',
        fontSize: '20px',
        color: selected ? 'LogoLight' : 'white',
        '&:hover': {
          color: 'NavMenuHover',
        },
        '&:active': {
          color: 'NavMenuActive',
        },
        '&:disabled': {
          color: 'LogoLight',
        },
      }}
      disabled={selected}
      disableRipple
    >
      <Box
        sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center' }}
      >
        {icon}
      </Box>
      <Box
        sx={{
          display: 'flex',
          flexDirection: 'row',
          width: '100%',
          alignItems: 'center',
          justifyContent: 'start',
        }}
      >
        {text}
        <Box sx={{ display: 'flex', flexGrow: 1, flexDirection: 'row' }} />
        {!!numUpdates && (
          <Box className={styles['notification-alert']}>
            <Typography
              variant="body2"
              sx={{ fontSize: '12px', fontWeight: 'light', color: 'white' }}
            >
              {numUpdates}
            </Typography>
          </Box>
        )}
      </Box>
    </Button>
  );
};

/**
 * The `MenuSidebar` is the core sidebar that we include throughout our UI. It
 * is pinned on the left-hand side of every page in our UI, and it includes
 * quick links to core abstractions in our system (workflows, integrations, etc).
 */
const MenuSidebar: React.FC<{ user: UserProfile }> = ({ user }) => {
  const dispatch: AppDispatch = useDispatch();
  const [currentPage, setCurrentPage] = useState(undefined);
  const location = useLocation();


  useEffect(() => {
    setCurrentPage(location.pathname);

    if (user) {
      dispatch(handleFetchNotifications({ user }));
    }
  }, []);

  const pathPrefix = getPathPrefix();
  const sidebarContent = (
    <>
      <Box className={styles['menu-sidebar-popover-container']}>
        <Link
          to={`${pathPrefix.length > 0 ? pathPrefix : '/'}`}
          underline="none"
          component={RouterLink as any}
        >
          <img
            src={
              'https://aqueduct-public-assets-bucket.s3.us-east-2.amazonaws.com/webapp/logos/aqueduct_logo_color_on_white.png'
            }
            width="50px"
            height="50px"
          />
        </Link>
      </Box>

      {/* <Box className={styles['menu-sidebar-links']}>

          <Link
            to={`${getPathPrefix()}/workflows`}
            className={styles['menu-sidebar-link']}
            underline="none"
            component={RouterLink as any}
          >
            <SidebarButton
              icon={
                <FontAwesomeIcon
                  className={styles['menu-sidebar-icon']}
                  icon={faShareNodes}
                />
              }
              text="Workflows"
              selected={currentPage === '/workflows'}
            />
          </Link>

          <Link
            to={`${getPathPrefix()}/integrations`}
            className={styles['menu-sidebar-link']}
            underline="none"
            component={RouterLink as any}
          >
            <SidebarButton
              icon={
                <FontAwesomeIcon
                  className={styles['menu-sidebar-icon']}
                  icon={faPlug}
                />
              }
              text="Integrations"
              selected={currentPage === '/integrations'}
            />
          </Link>

          <Link
            to={`${getPathPrefix()}/data`}
            className={styles['menu-sidebar-link']}
            underline="none"
            component={RouterLink as any}
          >
            <SidebarButton
              icon={
                <FontAwesomeIcon
                  className={styles['menu-sidebar-icon']}
                  icon={faDatabase}
                />
              }
              text="Data"
              selected={currentPage === '/data'}
            />
          </Link>
        </Box>

        <Box sx={{ width: '100%' }}>
          <Divider sx={{ width: '100%', backgroundColor: 'white' }} />
          <Box sx={{ my: 2 }}>
            <Link href="mailto:support@aqueducthq.com" underline="none">
              <SidebarButton
                icon={
                  <FontAwesomeIcon
                    className={styles['menu-sidebar-icon']}
                    icon={faMessage}
                  />
                }
                text="Report Issue"
              />
            </Link>
          </Box>
        </Box>
      </Box> */}
    </>
  );

  return <Box className={styles['menu-sidebar']}>{sidebarContent}</Box>;
};

export default MenuSidebar;
