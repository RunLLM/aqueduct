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
import { useNavigate } from 'react-router-dom';

import { theme } from '../../styles/theme/theme';
import ExecutionStatus from '../../utils/shared';
import { useSortedDagResults, useWorkflowIds } from '../pages/workflow/id/hook';

type Props = {
  apiKey: string;
};

export const VersionSelector: React.FC<Props> = ({ apiKey }) => {
  const navigate = useNavigate();
  const { workflowId, dagResultId } = useWorkflowIds(apiKey);

  const dagResults = useSortedDagResults(apiKey, workflowId);
  const selectedResult = (dagResults ?? []).filter(
    (r) => r.id === dagResultId
  )[0];
  console.log(dagResults)
  if (!dagResults || dagResults.length === 0) {
    return null;
  }
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
            navigate(
              `/workflow/${encodeURI(workflowId)}/result/${encodeURI(r.id)}`,
              { replace: false }
            );
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
            <Typography ml={1}>{`${new Date(
              r.exec_state.timestamps?.pending_at
            ).toLocaleString()}`}</Typography>
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
        {selectedResult && (
          <Box mx={1}>
            {new Date(
              selectedResult.exec_state.timestamps?.pending_at
            ).toLocaleString()}
          </Box>
        )}

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
