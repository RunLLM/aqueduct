import {
  faBook,
  faDatabase,
  faMessage,
  faPlug,
  faShareNodes,
} from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Link, Tooltip, Typography } from '@mui/material';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Divider from '@mui/material/Divider';
import React, { useEffect, useState } from 'react';
import { useDispatch } from 'react-redux';
import { Link as RouterLink, useLocation } from 'react-router-dom';

import { handleFetchNotifications } from '../../reducers/notifications';
import { AppDispatch } from '../../stores/store';
import UserProfile from '../../utils/auth';
import { getPathPrefix } from '../../utils/getPathPrefix';
import styles from './menu-sidebar-styles.module.css';

// Left padding = 8px
// Right padding = 8px
// Content size = 64px
export const MenuSidebarWidthNumber = 80;
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
        ...BUTTON_STYLE_OVERRIDE,
        bg: 'blue.800',
        fontSize: '10px',
        width: '64px',
        display: 'block',
        py: 1,
        px: 0,
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
      <Box>{icon}</Box>
      <Box
        sx={{
          marginTop: '8px',
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
  return (
    <Box className={styles['menu-sidebar']}>
      <Link
        to={`${pathPrefix.length > 0 ? pathPrefix : '/'}`}
        underline="none"
        className={styles['menu-sidebar-logo-link']}
        component={RouterLink}
      >
        <img
          src={
            'https://aqueduct-public-assets-bucket.s3.us-east-2.amazonaws.com/webapp/logos/aqueduct-logo-light/1x/logo_light_blue.png'
          }
          width="48px"
          height="48px"
        />
      </Link>

      <Box sx={{ my: 2 }} className={styles['menu-sidebar-content']}>
        <Tooltip title="Workflows" arrow placement="right">
          <Link
            to={`${getPathPrefix()}/workflows`}
            className={styles['menu-sidebar-link']}
            underline="none"
            component={RouterLink}
          >
            <SidebarButton
              icon={
                <FontAwesomeIcon
                  className={styles['menu-sidebar-icon']}
                  icon={faShareNodes}
                />
              }
              text=""
              selected={currentPage === '/workflows'}
            />
          </Link>
        </Tooltip>

        <Tooltip title="Integrations" arrow placement="right">
          <Link
            to={`${getPathPrefix()}/integrations`}
            className={styles['menu-sidebar-link']}
            underline="none"
            component={RouterLink}
          >
            <SidebarButton
              icon={
                <FontAwesomeIcon
                  className={styles['menu-sidebar-icon']}
                  icon={faPlug}
                />
              }
              text=""
              selected={currentPage === '/integrations'}
            />
          </Link>
        </Tooltip>

        <Tooltip title="Data" placement="right" arrow>
          <Link
            to={`${getPathPrefix()}/data`}
            className={styles['menu-sidebar-link']}
            underline="none"
            component={RouterLink}
          >
            <SidebarButton
              icon={
                <FontAwesomeIcon
                  className={styles['menu-sidebar-icon']}
                  icon={faDatabase}
                />
              }
              text=""
              selected={currentPage === '/data'}
            />
          </Link>
        </Tooltip>
      </Box>

      <Box className={styles['menu-sidebar-footer']}>
        <Divider sx={{ width: '100%', backgroundColor: 'white' }} />
        <Box sx={{ my: 2 }}>
          <Tooltip title="Documentation" placement="right" arrow>
            <Link href="https://docs.aqueducthq.com" underline="none">
              <SidebarButton
                icon={
                  <FontAwesomeIcon
                    className={styles['menu-sidebar-icon']}
                    icon={faBook}
                  />
                }
                text=""
              />
            </Link>
          </Tooltip>
        </Box>
        <Divider sx={{ width: '100%', backgroundColor: 'white' }} />
        <Box sx={{ my: 2 }}>
          <Tooltip title="Report Issue" placement="right" arrow>
            <Link href="mailto:support@aqueducthq.com" underline="none">
              <SidebarButton
                icon={
                  <FontAwesomeIcon
                    className={styles['menu-sidebar-icon']}
                    icon={faMessage}
                  />
                }
                text=""
              />
            </Link>
          </Tooltip>
        </Box>
      </Box>
    </Box>
  );
};

export default MenuSidebar;
