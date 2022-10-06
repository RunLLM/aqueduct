import { AppBar, Toolbar, Typography } from '@mui/material';
import Box from '@mui/material/Box';
import React from 'react';
import { theme } from '../../styles/theme/theme';

import UserProfile from '../../utils/auth';
import MenuSidebar, { MenuSidebarWidthNumber } from './menuSidebar';

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
      <Box sx={{ width: '100%', height: '100%', display: 'flex', flex: 1 }}>
        <MenuSidebar user={user} />

        {/* The margin here is fixed to be a constant (50px) more than the sidebar, which is a fixed width (200px). */}
        <Box
          sx={{
            marginLeft: MenuSidebarOffset,
            marginRight: 0,
            width: '100%',
            marginTop: 3,
          }}
        >
          {/* Header. +26 for the padding */}
          <AppBar sx={{
            width: `calc(100% - ${MenuSidebarWidthNumber + 26}px)`,
            boxShadow: 'none',
            borderBottom: `2px solid ${theme.palette.gray[300]}`,
            backgroundColor: 'white',
            color: 'black'
            }}>
            <Toolbar>
              <Typography variant="h6" component="div">
                Scroll to elevate App bar
              </Typography>
            </Toolbar>
          </AppBar>
          {children}
        </Box>
      </Box>
    </Box>
  );
};

export default DefaultLayout;
