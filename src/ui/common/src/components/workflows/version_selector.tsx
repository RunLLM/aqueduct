import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faClock } from '@fortawesome/free-solid-svg-icons';
import { buttonBaseClasses, InputLabel } from '@mui/material';
import Box from '@mui/material/Box';
import FormControl from '@mui/material/FormControl';
import MenuItem, { menuItemClasses } from '@mui/material/MenuItem';
import Select from '@mui/material/Select';
import Typography from '@mui/material/Typography';
import React from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';

import { selectResultIdx } from '../../reducers/workflow';
import { RootState } from '../../stores/store';
import { dateString } from '../../utils/metadata';
import { StatusIndicator } from './workflowStatus';
import { theme } from '../../styles/theme/theme';
import ExecutionStatus from '../../utils/shared';

export const VersionSelector: React.FC = () => {
  const navigate = useNavigate();
  const workflow = useSelector((state: RootState) => state.workflowReducer);
  const results = workflow.dagResults;
  const selectedResult = workflow.selectedResult;
  const dispatch = useDispatch();
  const [selectedResultIdx, setSelectedResultIdx] = React.useState(0);

  const getMenuItems = () => {
    return results.map((r, idx) => {
      const selected = selectedResult && selectedResult.id === r.id;

      if (selected && idx !== selectedResultIdx) {
        setSelectedResultIdx(idx);
      }

      const menuItemIcon = <StatusIndicator status={r.status} />;
      let color;

      if (r.status === ExecutionStatus.Succeeded) {
        color = theme.palette.green[600];
      } else if (r.status === ExecutionStatus.Failed) {
        color = theme.palette.red[600];
      } else {
        color = theme.palette.gray[700];
      }
      return (
        <MenuItem
          value={idx}
          key={r.id}
          onClick={() => {
            dispatch(selectResultIdx(idx));
            navigate(`?workflowDagResultId=${encodeURI(r.id)}`);
          }}
          sx={{ fontWeight: 'bold', color: color, }}
        >
          <Box sx={{ display: 'flex', alignItems: 'center' }}>
            <Box>{`${dateString(
              r.created_at
            )}`}</Box>
          </Box>
        </MenuItem>
      );
    });
  };

  return (
    <Box sx={{ display: 'flex', alignItems: 'center' }}>
      <FontAwesomeIcon icon={faClock} color={theme.palette.gray[700]} />
      <FormControl sx={{ minWidth: 120, ml: 1 }} size="small">
        <Select
          id="grouped-select"
          autoWidth
          variant="standard"
          value={selectedResultIdx}
          inputProps={{ startAdornment: <FontAwesomeIcon icon={faClock} color={theme.palette.gray[700]} /> }}
        >
          {getMenuItems()}
        </Select>
      </FormControl>
    </Box>
  );
};

export default VersionSelector;
