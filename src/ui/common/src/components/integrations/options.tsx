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
import React, {useEffect, useState} from 'react';

import { Integration, isDemo } from '../../utils/integrations';
import { Button } from '../primitives/Button.styles';
import {handleGetServerConfig} from "../../handlers/getServerConfig";

type Props = {
  integration: Integration;
  onUploadCsv?: () => void;
  onTestConnection?: () => void;
  onEdit?: () => void;
  onDeleteIntegration?: () => void;
  allowDeletion: boolean;
};

const IntegrationOptions: React.FC<Props> = ({
  integration,
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

  return (
    <Box display="flex" flexDirection="row" sx={{ height: 'fit-content' }}>
      {isDemo(integration) && (
        <Button
          variant="outlined"
          onClick={onUploadCsv}
          sx={{ width: '140px', marginRight: 1 }}
        >
          Upload CSV
        </Button>
      )}
      <Button
        color="primary"
        id={`options-${integration.id}`}
        onClick={(event) => {
          setAnchorEl(event.currentTarget);
        }}
        endIcon={<FontAwesomeIcon icon={faCaretDown} size="sm" />}
        sx={{ width: '120px' }}
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

        {!isDemo(integration) && (
          <MenuItem
            onClick={() => {
              setAnchorEl(null);
              onEdit();
            }}
          >
            <FontAwesomeIcon color="gray.800" icon={faPen} width="16px" />
            <Typography color="gray.800" variant="body2" sx={{ marginLeft: 1 }}>
              Edit Integration
            </Typography>
          </MenuItem>
        )}
        {!isDemo(integration) && allowDeletion && (
          <MenuItem
            onClick={() => {
              setAnchorEl(null);
              onDeleteIntegration();
            }}
          >
            <FontAwesomeIcon color="gray.800" icon={faTrash} />
            <Typography color="gray.800" variant="body2" sx={{ marginLeft: 1 }}>
              Delete Integration
            </Typography>
          </MenuItem>
        )}
      </Menu>
    </Box>
  );
};

export default IntegrationOptions;
