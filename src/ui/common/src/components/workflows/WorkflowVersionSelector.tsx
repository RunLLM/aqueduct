import {
  faChevronDown,
  faCircleCheck,
  faCircleXmark,
  faClock,
  faSpinner,
} from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Popover } from '@mui/material';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import MenuItem from '@mui/material/MenuItem';
import Typography from '@mui/material/Typography';
import React, { useState } from 'react';
import { useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';
import { useDagResultsGetQuery } from 'src/handlers/AqueductApi';

import { RootState } from '../../stores/store';
import { theme } from '../../styles/theme/theme';
import ExecutionStatus from '../../utils/shared';

type Props = {
  apiKey: string;
  workflowId: string;
};

export const VersionSelector: React.FC<Props> = ({ apiKey, workflowId }) => {
  const navigate = useNavigate();
  const pageState = useSelector(
    (state: RootState) =>
      state.workflowPageReducer.perWorkflowPageStates[workflowId]
  );
  const dagResultId = pageState?.dagResultId;

  const { data: dagResults } = useDagResultsGetQuery({ apiKey, workflowId });
  const [menuAnchor, setMenuAnchor] = useState<HTMLButtonElement | null>(null);

  const getMenuItems = () => {
    if (!dagResults || dagResults.length === 0) {
      return [];
    }
    return dagResults.map((r, idx) => {
      // either an ID match, or no selection and default to the first result
      const selected = r.id === dagResultId || (!dagResultId && idx === 0);

      let menuItemIcon;

      let defaultBackground, hoverBackground, selectedBackground;

      switch (r.exec_state.status) {
        case ExecutionStatus.Succeeded:
          defaultBackground = theme.palette.green[100];
          hoverBackground = theme.palette.green[25];
          selectedBackground = theme.palette.green[200];

          menuItemIcon = (
            <Box sx={{ color: theme.palette.Success }}>
              <FontAwesomeIcon icon={faCircleCheck} />
            </Box>
          );
          break;
        case ExecutionStatus.Pending:
          defaultBackground = theme.palette.gray[100];
          hoverBackground = theme.palette.gray[25];
          selectedBackground = theme.palette.gray[200];

          menuItemIcon = (
            <Box sx={{ color: theme.palette.gray['700'] }}>
              <FontAwesomeIcon icon={faSpinner} spin={true} />
            </Box>
          );
          break;
        case ExecutionStatus.Failed:
          defaultBackground = theme.palette.red[100];
          hoverBackground = theme.palette.red[25];
          selectedBackground = theme.palette.red[300];

          menuItemIcon = (
            <Box sx={{ color: theme.palette.Error }}>
              <FontAwesomeIcon icon={faCircleXmark} />
            </Box>
          );
          break;
      }

      return (
        <MenuItem
          value={idx}
          key={r.id}
          onClick={() => {
            navigate(`?workflowDagId=${encodeURI(r.dag_id)}&workflowDagResultId=${encodeURI(r.id)}`);
          }}
          sx={{
            backgroundColor: selected ? selectedBackground : defaultBackground,
            ':hover': {
              backgroundColor: hoverBackground,
            },
          }}
          disableRipple
        >
          <Box sx={{ display: 'flex', alignItems: 'center' }}>
            {menuItemIcon}
            <Typography
              ml={1}
            >{`${r.exec_state.timestamps?.pending_at}`}</Typography>
          </Box>
        </MenuItem>
      );
    });
  };

  return (
    <Box sx={{ display: 'flex', alignItems: 'center' }}>
      <Button
        sx={{
          backgroundColor: !!menuAnchor ? theme.palette.gray[50] : 'white',
          ':hover': {
            backgroundColor: theme.palette.gray[50],
          },
          p: 1,
          borderRadius: '4px',
          fontSize: '16px',
          display: 'flex',
          alignItems: 'center',
          textTransform: 'none',
          color: theme.palette.gray[900],
        }}
        onClick={(e) => setMenuAnchor(e.currentTarget)}
        disableRipple
        disableFocusRipple
      >
        <FontAwesomeIcon icon={faClock} color={theme.palette.gray[800]} />
        <Box mx={1}>{r.exec_state.timestamps?.pending_at}</Box>

        <FontAwesomeIcon icon={faChevronDown} />
      </Button>

      <Popover
        open={!!menuAnchor}
        anchorEl={menuAnchor}
        onClose={() => setMenuAnchor(null)}
        anchorOrigin={{
          vertical: 'bottom',
          horizontal: 'left',
        }}
      >
        {getMenuItems()}
      </Popover>
    </Box>
  );
};

export default VersionSelector;
