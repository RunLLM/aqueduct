import {
  faCircleCheck,
  faCircleXmark,
  faSpinner,
} from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import Box from '@mui/material/Box';
import FormControl from '@mui/material/FormControl';
import MenuItem from '@mui/material/MenuItem';
import Select from '@mui/material/Select';
import Typography from '@mui/material/Typography';
import React from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { setAllSideSheetState } from '../../reducers/openSideSheet';
import { selectResultIdx } from '../../reducers/workflow';
import { RootState } from '../../stores/store';
import { theme } from '../../styles/theme/theme';
import { dateString } from '../../utils/metadata';
import ExecutionStatus from '../../utils/shared';

export const VersionSelector: React.FC = () => {
  const workflow = useSelector((state: RootState) => state.workflowReducer);
  const results = workflow.dagResults;
  const selectedResult = workflow.selectedResult;
  const dispatch = useDispatch();
  const [selectedResultIdx, setSelectedResultIdx] = React.useState(0);

  const menuItems = results.map((r, idx) => {
    const selected = selectedResult && selectedResult.id === r.id;

    if (selected && idx !== selectedResultIdx) {
      setSelectedResultIdx(idx);
    }

    let menuItemIcon;
    switch (r.status) {
      case ExecutionStatus.Succeeded:
        menuItemIcon = (
          <Box sx={{ fontSize: '20px', color: theme.palette.green['500'] }}>
            <FontAwesomeIcon icon={faCircleCheck} />
          </Box>
        );
        break;
      case ExecutionStatus.Pending:
        menuItemIcon = (
          <Box sx={{ fontSize: '20px', color: theme.palette.gray['700'] }}>
            <FontAwesomeIcon icon={faSpinner} />
          </Box>
        );
        break;
      case ExecutionStatus.Failed:
        menuItemIcon = (
          <Box sx={{ fontSize: '20px', color: theme.palette.red['500'] }}>
            <FontAwesomeIcon icon={faCircleXmark} />
          </Box>
        );
        break;
    }

    return (
      <MenuItem
        value={idx}
        key={idx}
        onClick={() => {
          dispatch(setAllSideSheetState(false));
          dispatch(selectResultIdx(idx));
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

  return (
    <Box sx={{ display: 'flex', alignItems: 'center' }}>
      <FormControl sx={{ m: 1, minWidth: 120 }} size="small">
        <Select
          sx={{ maxHeight: 50 }}
          id="grouped-select"
          autoWidth
          value={selectedResultIdx}
        >
          {menuItems}
        </Select>
      </FormControl>
    </Box>
  );
};

export default VersionSelector;
