import { createTheme, ThemeProvider } from '@mui/material';
import Box from '@mui/material/Box';
import Link from '@mui/material/Link';
import Typography from '@mui/material/Typography';
import React, { useEffect } from 'react';
import { useDispatch } from 'react-redux';

import { handleFetchNotifications } from '../../reducers/notifications';
import { AppDispatch } from '../../stores/store';
import { theme } from '../../styles/theme/theme';
import UserProfile from '../../utils/auth';
import { breadcrumbsSize } from '../notifications/NotificationsPopover';
import MenuSidebar, { MenuSidebarWidth } from './menuSidebar';
import NavBar, { BreadcrumbLink } from './NavBar';
import {FontAwesomeIcon} from "@fortawesome/react-fontawesome";
import {faXmark} from "@fortawesome/free-solid-svg-icons";

export const DefaultLayoutMargin = '24px';
export const SidesheetMargin = '16px';
export const SidesheetWidth = '800px';
export const SidesheetContentWidth = '768px'; // 800 - 16 - 16
export const SidesheetButtonHeight = '40px';

type Props = {
  user: UserProfile;
  children: React.ReactElement | React.ReactElement[];
  breadcrumbs: BreadcrumbLink[];
  /**
   * Function to be called when breadcrumbs are clicked. Useful for doing cleanup on navigation.
   */
  onBreadCrumbClicked?: (name: string) => void;
  onSidebarItemClicked?: (name: string) => void;
};

export const DefaultLayout: React.FC<Props> = ({
  user,
  children,
  breadcrumbs,
  onBreadCrumbClicked = null,
  onSidebarItemClicked = null,
}) => {
  const muiTheme = createTheme(theme);
  const dispatch: AppDispatch = useDispatch();

  useEffect(() => {
    if (user) {
      dispatch(handleFetchNotifications({ user }));
      // TODO: Get the version number here.
      // Store version number after it's fetched.
      // lastVersionNumberSeen, showBanner are two vars that we should use to track whether or not to show the verison number
    }
  }, [dispatch, user]);

  return (
    <ThemeProvider theme={muiTheme}>
      <Box
        sx={{
          position: 'relative',
          zIndex: 999,
          backgroundColor: '#A7E2EA',
          width: '100%',
          height: '64px',
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
        }}
      >
        <Box>
          <Typography variant="h6">
            ✨v0.2.8 has launched!{' '}
            <Link href={'https://github.com/aqueducthq/aqueduct/releases'}>Release Notes</Link>
          </Typography>
        </Box>
        <Box sx={{ width: '16px', fontSize: '16px', display: 'flex', alignItems: 'center', justifySelf: 'space-between', position: 'absolute', right: '16px' }}>
          <FontAwesomeIcon icon={faXmark} />
        </Box>
      </Box>

      <Box
        sx={{
          width: '100%',
          height: '100%',
          position: 'relative',
          overflow: 'auto',
          top: '64px',
        }}
      >
        <Box sx={{ width: '100%', height: '100%', display: 'flex', flex: 1 }}>
          <MenuSidebar
            user={user}
            onSidebarItemClicked={onSidebarItemClicked}
          />
          <NavBar
            user={user}
            breadcrumbs={breadcrumbs}
            onBreadCrumbClicked={onBreadCrumbClicked}
          />
          {/* Pad top for breadcrumbs (64px). */}
          {/* The margin here is fixed to be a constant (50px) more than the sidebar, which is a fixed width (200px). */}
          <Box
            sx={{
              boxSizing: 'border-box',
              width: `calc(100% - ${MenuSidebarWidth} - ${DefaultLayoutMargin})`,
              marginTop: breadcrumbsSize,
              marginLeft: MenuSidebarWidth,
              marginRight: 0,
              paddingTop: DefaultLayoutMargin,
              paddingLeft: DefaultLayoutMargin,
            }}
          >
            {children}
          </Box>
        </Box>
      </Box>
    </ThemeProvider>
  );
};

export default DefaultLayout;
