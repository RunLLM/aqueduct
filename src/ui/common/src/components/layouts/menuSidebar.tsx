import {
  faBook,
  faDatabase,
  faMessage,
  faPlug,
  faShareNodes,
} from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Link, Typography } from '@mui/material';
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
import { useEnvironmentGetQuery } from '../../handlers/AqueductApi';

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
          marginTop: '4px',
          fontSize: '12px',
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

  const { data } = useEnvironmentGetQuery({ apiKey: user.apiKey } as any, { skip: !user?.apiKey });

  useEffect(() => {
    setCurrentPage(location.pathname);
  }, [dispatch, location.pathname]);

  useEffect(() => {
    if (data) {
      setVersionNumber(data.version);
    }
  }, [user.apiKey, data]);

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
            text="Workflows"
            selected={currentPage === '/workflows'}
          />
        </Link>

        <Divider sx={{ width: '64px', backgroundColor: 'white' }} />

        <Link
          to={`${getPathPrefix()}/resources`}
          style={menuSidebarLink}
          underline="none"
          component={RouterLink}
        >
          <SidebarButton
            onClick={() => {
              if (onSidebarItemClicked) {
                onSidebarItemClicked('resources');
              }
            }}
            icon={<FontAwesomeIcon style={menuSidebarIcon} icon={faPlug} />}
            text="Resources"
            selected={currentPage === '/resources'}
          />
        </Link>

        <Divider sx={{ width: '64px', backgroundColor: 'white' }} />

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
            icon={<FontAwesomeIcon style={menuSidebarIcon} icon={faDatabase} />}
            text="Data"
            selected={currentPage === '/data'}
          />
        </Link>

        <Divider sx={{ width: '64px', backgroundColor: 'white' }} />
      </Box>

      <Box style={menuSidebarFooter}>
        <Divider sx={{ width: '64px', backgroundColor: 'white' }} />
        <Box sx={{ my: 2 }}>
          <Link
            href="https://docs.aqueducthq.com"
            underline="none"
            target="_blank"
            rel="noreferrer noopener"
          >
            <SidebarButton
              onClick={() => {
                if (onSidebarItemClicked) {
                  onSidebarItemClicked('documentation');
                }
              }}
              icon={<FontAwesomeIcon style={menuSidebarIcon} icon={faBook} />}
              text="Docs"
            />
          </Link>
        </Box>
        <Divider sx={{ width: '64px', backgroundColor: 'white' }} />
        <Box sx={{ my: 2 }}>
          <Link
            href="mailto:support@aqueducthq.com"
            underline="none"
            target="_blank"
            rel="noreferrer noopener"
          >
            <SidebarButton
              onClick={() => {
                if (onSidebarItemClicked) {
                  onSidebarItemClicked('report_issue');
                }
              }}
              icon={
                <FontAwesomeIcon style={menuSidebarIcon} icon={faMessage} />
              }
              text="Feedback"
            />
          </Link>
        </Box>
        <Box marginLeft="14px" marginBottom="16px">
          <Typography variant="caption" sx={{ color: 'white' }}>
            {versionNumber.length > 0 ? `v${versionNumber}` : ''}
          </Typography>
        </Box>
      </Box>
    </Box>
  );
};

export default MenuSidebar;
