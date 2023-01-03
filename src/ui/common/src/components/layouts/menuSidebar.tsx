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
import UserProfile from 'src/utils/auth';

import { AppDispatch } from '../../stores/store';
import { getPathPrefix } from '../../utils/getPathPrefix';
import { apiAddress } from '../hooks/useAqueductConsts';
import {
  menuSidebar,
  menuSidebarContent,
  menuSidebarFooter,
  menuSidebarIcon,
  menuSidebarLink,
  menuSidebarLogoLink,
  notificationAlert,
} from './menuSidebar.styles';

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
  onClick?: () => void;
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
  onClick,
}) => {
  return (
    <Button
      onClick={onClick}
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
          <Box style={notificationAlert}>
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
const MenuSidebar: React.FC<{
  onSidebarItemClicked?: (name: string) => void;
  user: UserProfile;
}> = ({ onSidebarItemClicked, user }) => {
  const dispatch: AppDispatch = useDispatch();
  const [currentPage, setCurrentPage] = useState(undefined);
  const [versionNumber, setVersionNumber] = useState('');
  const location = useLocation();

  useEffect(() => {
    setCurrentPage(location.pathname);
  }, [dispatch, location.pathname]);

  useEffect(() => {
    async function fetchVersionNumber() {
      const res = await fetch(`${apiAddress}/api/version`, { method: 'GET', headers: { 'api-key': user.apiKey } });
      const versionNumberResponse = await res.json();

      if (!res.ok) {
        console.log('error getting version number', versionNumberResponse.error);
      }

      console.log('versionNumberResponse: ', versionNumberResponse);
      setVersionNumber(versionNumberResponse.version);
    }

    fetchVersionNumber();
  }, [])

  const pathPrefix = getPathPrefix();
  return (
    <Box style={menuSidebar}>
      <Link
        to={`${pathPrefix.length > 0 ? pathPrefix : '/'}`}
        underline="none"
        style={menuSidebarLogoLink}
        component={RouterLink}
        onClick={() => {
          if (onSidebarItemClicked) {
            onSidebarItemClicked('home');
          }
        }}
      >
        <img
          src={
            'https://aqueduct-public-assets-bucket.s3.us-east-2.amazonaws.com/webapp/logos/aqueduct-logo-light/1x/logo_light_blue.png'
          }
          width="48px"
          height="48px"
        />
      </Link>

      <Box sx={{ my: 2 }} style={menuSidebarContent}>
        <Tooltip title="Workflows" arrow placement="right">
          <Link
            to={`${getPathPrefix()}/workflows`}
            style={menuSidebarLink}
            underline="none"
            component={RouterLink}
          >
            <SidebarButton
              onClick={() => {
                if (onSidebarItemClicked) {
                  onSidebarItemClicked('workflows');
                }
              }}
              icon={
                <FontAwesomeIcon style={menuSidebarIcon} icon={faShareNodes} />
              }
              text=""
              selected={currentPage === '/workflows'}
            />
          </Link>
        </Tooltip>

        <Tooltip title="Integrations" arrow placement="right">
          <Link
            to={`${getPathPrefix()}/integrations`}
            style={menuSidebarLink}
            underline="none"
            component={RouterLink}
          >
            <SidebarButton
              onClick={() => {
                if (onSidebarItemClicked) {
                  onSidebarItemClicked('integrations');
                }
              }}
              icon={<FontAwesomeIcon style={menuSidebarIcon} icon={faPlug} />}
              text=""
              selected={currentPage === '/integrations'}
            />
          </Link>
        </Tooltip>

        <Tooltip title="Data" placement="right" arrow>
          <Link
            to={`${getPathPrefix()}/data`}
            style={menuSidebarLink}
            underline="none"
            component={RouterLink}
          >
            <SidebarButton
              onClick={() => {
                if (onSidebarItemClicked) {
                  onSidebarItemClicked('data');
                }
              }}
              icon={
                <FontAwesomeIcon style={menuSidebarIcon} icon={faDatabase} />
              }
              text=""
              selected={currentPage === '/data'}
            />
          </Link>
        </Tooltip>
      </Box>

      <Box style={menuSidebarFooter}>
        <Divider sx={{ width: '64px', backgroundColor: 'white' }} />
        <Box sx={{ my: 2 }}>
          <Tooltip title="Documentation" placement="right" arrow>
            <Link href="https://docs.aqueducthq.com" underline="none">
              <SidebarButton
                onClick={() => {
                  if (onSidebarItemClicked) {
                    onSidebarItemClicked('documentation');
                  }
                }}
                icon={<FontAwesomeIcon style={menuSidebarIcon} icon={faBook} />}
                text=""
              />
            </Link>
          </Tooltip>
        </Box>
        <Divider sx={{ width: '64px', backgroundColor: 'white' }} />
        <Box sx={{ my: 2 }}>
          <Tooltip title="Report Issue" placement="right" arrow>
            <Link href="mailto:support@aqueducthq.com" underline="none">
              <SidebarButton
                onClick={() => {
                  if (onSidebarItemClicked) {
                    onSidebarItemClicked('report_issue');
                  }
                }}
                icon={
                  <FontAwesomeIcon style={menuSidebarIcon} icon={faMessage} />
                }
                text=""
              />
            </Link>
          </Tooltip>
        </Box>
        <Box marginLeft="16px">
          <Typography variant="caption" sx={{ color: 'white' }}>v{versionNumber}</Typography>
        </Box>
      </Box>
    </Box>
  );
};

export default MenuSidebar;
