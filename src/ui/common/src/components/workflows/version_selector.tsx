import { InputLabel } from '@mui/material';
import Box from '@mui/material/Box';
import FormControl from '@mui/material/FormControl';
import MenuItem from '@mui/material/MenuItem';
import Select from '@mui/material/Select';
import Typography from '@mui/material/Typography';
import React from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';

import { selectResultIdx } from '../../reducers/workflow';
import { RootState } from '../../stores/store';
import { dateString } from '../../utils/metadata';
import { StatusIndicator } from './workflowStatus';

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
      return (
        <MenuItem
          value={idx}
          key={r.id}
          onClick={() => {
            dispatch(selectResultIdx(idx));
            navigate(`?workflowDagResultId=${encodeURI(r.id)}`);
          }}
          sx={{ backgroundColor: selected ? 'blueTint' : null }}
        >
          <Box sx={{ display: 'flex', alignItems: 'center' }}>
            {menuItemIcon}
            <Typography sx={{ ml: 1 }}>{`${dateString(
              r.created_at
            )}`}</Typography>
          </Box>
        </MenuItem>
      );
    });
  };

  return (
    <Box sx={{ display: 'flex', alignItems: 'center' }}>
      <FormControl sx={{ minWidth: 120 }} size="small">
        <InputLabel id="version-label">Version</InputLabel>
        <Select
          id="grouped-select"
          autoWidth
          label="Version"
          value={selectedResultIdx}
        >
          {getMenuItems()}
        </Select>
      </FormControl>
    </Box>
  );
};

export default VersionSelector;
