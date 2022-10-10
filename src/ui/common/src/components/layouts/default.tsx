import Box from '@mui/material/Box';
import React from 'react';

import UserProfile from '../../utils/auth';
import { breadcrumbsSize } from '../notifications/NotificationsPopover';
import MenuSidebar, { MenuSidebarWidthNumber } from './menuSidebar';
import NavBar, { BreadcrumbLinks } from './NavBar';

export const MenuSidebarOffset = `${MenuSidebarWidthNumber + 50}px`;

type Props = {
  user: UserProfile;
  children: React.ReactElement | React.ReactElement[];
  breadcrumbs: BreadcrumbLinks[];
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
            paddingTop: breadcrumbsSize,
            marginLeft: MenuSidebarOffset,
            marginRight: 0,
            width: '100%',
            marginTop: 3,
          }}
        >
          {children}
        </Box>
      </Box>
    </Box>
  );
};

export default DefaultLayout;
