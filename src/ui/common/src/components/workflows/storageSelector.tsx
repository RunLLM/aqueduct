
import React from 'react';
import Box from '@mui/material/Box';
import FormControl from '@mui/material/FormControl';
import MenuItem from '@mui/material/MenuItem';
import Select from '@mui/material/Select';
import Typography from '@mui/material/Typography';
import { useSelector } from 'react-redux';
import { StorageType, StorageTypeNames } from '../../utils/storage';

import { RootState } from '../../stores/store';

export const StorageSelector: React.FC = () => {
  const workflow = useSelector((state: RootState) => state.workflowReducer);
  const dag = workflow.selectedDag;
  let selected = "file";
  if (dag) {
    selected = dag.storage_config.type;
    console.log("selected dag", dag.storage_config.type);
  }
  console.log(dag);

  const getMenuItems = () => {
    return (Object.values(StorageType) as Array<typeof StorageType[keyof typeof StorageType]>).map((r, _) => {
      return (
        <MenuItem
          value={r}
          key={r}
          sx={{ backgroundColor: selected ? 'blueTint' : null }}
        >
          <Box sx={{ display: 'flex', alignItems: 'center' }}>
            <Typography sx={{ ml: 1 }}>{StorageTypeNames[r]}</Typography>
          </Box>
        </MenuItem>
      );
    });
  };

  let menuItems = getMenuItems();

  return (
    <Box sx={{ display: 'flex', alignItems: 'center' }}>
      Storage Type:{' '}
      <FormControl disabled sx={{ m: 1, minWidth: 120 }} size="small">
        <Select
          sx={{ maxHeight: 50 }}
          id="grouped-select"
          autoWidth
          value={selected}
        >
          {menuItems}
        </Select>
      </FormControl>
    </Box>
  );
};

export default StorageSelector;
