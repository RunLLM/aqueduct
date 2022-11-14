import Box from '@mui/material/Box';
import React from 'react';

import UserProfile from '../../utils/auth';
import { breadcrumbsSize } from '../notifications/NotificationsPopover';
import MenuSidebar, { MenuSidebarWidth } from './menuSidebar';
import NavBar, { BreadcrumbLink } from './NavBar';

//export const MenuSidebarOffset = `${MenuSidebarWidthNumber + 50}px`;
export const DefaultLayoutMargin = '24px';
export const SidesheetMargin = '16px';
export const SidesheetWidth = '800px';
export const SidesheetContentWidth = '768px'; // 800 - 16 - 16
export const SidesheetButtonHeight = '40px';

type Props = {
  user: UserProfile;
  children: React.ReactElement | React.ReactElement[];
  breadcrumbs: BreadcrumbLink[];
};

export const DefaultLayout: React.FC<Props> = ({
  user,
  children,
  breadcrumbs,
}) => {
  return (
    <Box
      sx={{
        width: '100%',
        height: '100%',
        position: 'fixed',
        overflow: 'auto',
      }}
    >
      <Box sx={{ width: '100%', height: '100%', display: 'flex', flex: 1 }}>
        <MenuSidebar user={user} />
        <NavBar user={user} breadcrumbs={breadcrumbs} />
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
  );
};

export default DefaultLayout;
