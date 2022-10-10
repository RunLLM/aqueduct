import { AppBar, Breadcrumbs, Link, Toolbar, Typography } from '@mui/material';
import Box from '@mui/material/Box';
import React from 'react';
import { theme } from '../../styles/theme/theme';
import { Link as RouterLink } from 'react-router-dom';

import UserProfile from '../../utils/auth';
import MenuSidebar, { MenuSidebarWidthNumber } from './menuSidebar';
import NavBar from './NavBar';

export const MenuSidebarOffset = `${MenuSidebarWidthNumber + 50}px`;

type Props = {
  user: UserProfile;
  children: React.ReactElement | React.ReactElement[];
};

export const DefaultLayout: React.FC<Props> = ({ user, children }) => {
  return (
    <Box
      sx={{
        width: '100%',
        height: '100%',
        position: 'fixed',
        overflow: 'auto',
      }}
    >
      <Box sx={{ width: '100%', height: '100%', display: 'flex', flex: 1}}>

        <MenuSidebar user={user} />

        <NavBar user={user} />
        {/* Pad top for breadcrumbs (64px). */}
        {/* The margin here is fixed to be a constant (50px) more than the sidebar, which is a fixed width (200px). */}
        <Box
          sx={{
            paddingTop: '64px',
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
