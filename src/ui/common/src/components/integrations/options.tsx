<<<<<<< HEAD
import { faCaretDown, faFlask, faPen } from '@fortawesome/free-solid-svg-icons';
=======
import {
  faCaretDown,
  faFlask,
  faPen,
  faTrash,
} from '@fortawesome/free-solid-svg-icons';
>>>>>>> 2640da16405aadfa794e87c26503038fb16765fe
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import Box from '@mui/material/Box';
import Menu from '@mui/material/Menu';
import MenuItem from '@mui/material/MenuItem';
import Typography from '@mui/material/Typography';
import React, { useState } from 'react';
import { useSelector } from 'react-redux';

import { RootState } from '../../stores/store';
import { Integration, isDemo } from '../../utils/integrations';
import { LoadingStatusEnum } from '../../utils/shared';
import { Button } from '../primitives/Button.styles';

type Props = {
  integration: Integration;
  onUploadCsv?: () => void;
  onTestConnection?: () => void;
  onEdit?: () => void;
<<<<<<< HEAD
=======
  onDeleteIntegration?: () => void;
>>>>>>> 2640da16405aadfa794e87c26503038fb16765fe
};

const IntegrationOptions: React.FC<Props> = ({
  integration,
  onUploadCsv,
  onTestConnection,
  onEdit,
<<<<<<< HEAD
=======
  onDeleteIntegration,
>>>>>>> 2640da16405aadfa794e87c26503038fb16765fe
}) => {
  // Menu control based on
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const onMenuClose = () => {
    setAnchorEl(null);
  };
  const operatorsState = useSelector((state: RootState) => {
    return state.integrationReducer.operators;
  });
  let inUse = true;
  if (
    operatorsState.status.loading === LoadingStatusEnum.Succeeded &&
    operatorsState.operators.length === 0
  ) {
    inUse = false;
  }
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
<<<<<<< HEAD
=======

>>>>>>> 2640da16405aadfa794e87c26503038fb16765fe
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
        {!isDemo(integration) && (
          <MenuItem
            onClick={() => {
              setAnchorEl(null);
              onDeleteIntegration();
            }}
            disabled={inUse}
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
