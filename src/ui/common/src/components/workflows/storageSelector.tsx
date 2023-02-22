import { Link } from '@mui/material';
import Box from '@mui/material/Box';
import FormControl from '@mui/material/FormControl';
import MenuItem from '@mui/material/MenuItem';
import Select from '@mui/material/Select';
import Typography from '@mui/material/Typography';
import React from 'react';
import { useSelector } from 'react-redux';

import { RootState } from '../../stores/store';
import { StorageType, StorageTypeNames } from '../../utils/storage';

export const StorageSelector: React.FC = () => {
  const workflow = useSelector((state: RootState) => state.workflowReducer);
  const dag = workflow.selectedDag;
  let selected = 'file';
  let selectedLocation = null;
  if (dag) {
    console.log('dag: ', dag);
    selected = dag.storage_config.type;
    switch (selected) {
      case 's3': {
        selectedLocation = dag.storage_config.s3_config.bucket;
        break;
      }
      case 'gcs': {
        selectedLocation = dag.storage_config.gcs_config.bucket;
        break;
      }
      case 'file':
      default: {
        selectedLocation = dag.storage_config.file_config.directory;
        break;
      }
    }
  }

  const getMenuItems = () => {
    return (
      Object.values(StorageType) as Array<
        (typeof StorageType)[keyof typeof StorageType]
      >
    ).map((storageType) => {
      return (
        <MenuItem
          value={storageType}
          key={storageType}
          sx={{ backgroundColor: selected ? 'blueTint' : null }}
        >
          <Box sx={{ display: 'flex', alignItems: 'center' }}>
            <Typography>{StorageTypeNames[storageType]}</Typography>
          </Box>
        </MenuItem>
      );
    });
  };

  const menuItems = getMenuItems();

  return (
    <Box>
      <Box sx={{ display: 'flex' }}>
        <Typography style={{ fontWeight: 'bold' }}>
          Metadata Storage Location:
        </Typography>
        <Typography sx={{ marginLeft: '8px' }}>{selectedLocation}</Typography>
      </Box>
      <Typography variant="body2">
        For more details on modifying the Aqueduct metadata store, please see{' '}
        <Link href="https://docs.aqueducthq.com/guides/changing-metadata-store">
          our documentation
        </Link>
        .
      </Typography>
      <FormControl disabled sx={{ minWidth: 120 }} size="small">
        <Select
          sx={{ maxHeight: 48 }}
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
