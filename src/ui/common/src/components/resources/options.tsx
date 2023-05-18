import {
  faCaretDown,
  faFlask,
  faPen,
  faTrash,
} from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import Box from '@mui/material/Box';
import Menu from '@mui/material/Menu';
import MenuItem from '@mui/material/MenuItem';
import Typography from '@mui/material/Typography';
import React, { useState } from 'react';

import {
  Integration,
  isBuiltinIntegration,
  isCondaRegistered,
  resourceExecState,
} from '../../utils/resources';
import ExecutionStatus from '../../utils/shared';
import { Button } from '../primitives/Button.styles';

type Props = {
  resource: Integration;

  // Currently unused.
  onUploadCsv?: () => void;
  onTestConnection?: () => void;
  onEdit?: () => void;
  onDeleteIntegration?: () => void;
  allowDeletion: boolean;
};

export const IntegrationOptionsButtonWidth = '120px';

const IntegrationOptions: React.FC<Props> = ({
  resource,
  onUploadCsv,
  onTestConnection,
  onEdit,
  onDeleteIntegration,
  allowDeletion,
}) => {
  // Menu control based on
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const onMenuClose = () => {
    setAnchorEl(null);
  };

  // Disallow any deletion for the built-in resources, unless Conda has completed registration.
  let deletionMenuItem = 'Delete Resource';
  if (isBuiltinIntegration(resource)) {
    allowDeletion = false;
  }

  if (
    resource.service === 'Aqueduct' &&
    isCondaRegistered(resource) &&
    (resourceExecState(resource).status === ExecutionStatus.Succeeded ||
      resourceExecState(resource).status === ExecutionStatus.Failed)
  ) {
    allowDeletion = true;
    deletionMenuItem = 'Delete Conda';
  }

  return (
    <Box display="flex" flexDirection="row" sx={{ height: 'fit-content' }}>
      <Button
        color="primary"
        id={`options-${resource.id}`}
        onClick={(event) => {
          setAnchorEl(event.currentTarget);
        }}
        endIcon={<FontAwesomeIcon icon={faCaretDown} size="sm" />}
        sx={{ width: { IntegrationOptionsButtonWidth } }}
      >
        Options
      </Button>
      <Menu
        elevation={1}
        open={!!anchorEl}
        sx={{ marginTop: 1 }}
        anchorEl={anchorEl}
        onClose={onMenuClose}
        // These two fields controls positioning and alignment of the menu
        // w.r.t. the button. https://mui.com/material-ui/react-popover/#anchor-playground
        anchorOrigin={{
          vertical: 'bottom',
          horizontal: 'right',
        }}
        transformOrigin={{
          vertical: 'top',
          horizontal: 'right',
        }}
      >
        <MenuItem
          onClick={() => {
            setAnchorEl(null);
            onTestConnection();
          }}
        >
          <FontAwesomeIcon color="gray.800" icon={faFlask} width="16px" />
          <Typography color="gray.800" variant="body2" sx={{ marginLeft: 1 }}>
            Test Connection
          </Typography>
        </MenuItem>

        {resource.service !== 'AWS' &&
          !isBuiltinIntegration(resource) && (
            <MenuItem
              onClick={() => {
                setAnchorEl(null);
                onEdit();
              }}
            >
              <FontAwesomeIcon color="gray.800" icon={faPen} width="16px" />
              <Typography
                color="gray.800"
                variant="body2"
                sx={{ marginLeft: 1 }}
              >
                Edit Resource
              </Typography>
            </MenuItem>
          )}
        {allowDeletion && (
          <MenuItem
            onClick={() => {
              setAnchorEl(null);
              onDeleteIntegration();
            }}
          >
            <FontAwesomeIcon color="gray.800" icon={faTrash} />
            <Typography color="gray.800" variant="body2" sx={{ marginLeft: 1 }}>
              {deletionMenuItem}
            </Typography>
          </MenuItem>
        )}
      </Menu>
    </Box>
  );
};

export default IntegrationOptions;
