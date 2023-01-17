import Box from '@mui/material/Box';
import FormControl from '@mui/material/FormControl';
import MenuItem from '@mui/material/MenuItem';
import Select from '@mui/material/Select';
import Typography from '@mui/material/Typography';
import React from 'react';
import { ListWorkflowSummary } from 'src/utils/workflows';

type Props = {
  sourceId: string;
  setSourceId: (string) => void;
  workflows: ListWorkflowSummary[];
};

export const TriggerSourceSelector: React.FC<Props> = ({
  sourceId,
  setSourceId,
  workflows,
}) => {
  const getMenuItems = () => {
    return workflows.map((workflow) => {
      return (
        <MenuItem
          key={workflow.id}
          value={workflow.id}
          sx={{ backgroundColor: 'blueTint' }}
        >
          <Box sx={{ display: 'flex', alignItems: 'center' }}>
            <Typography>{workflow.name}</Typography>
          </Box>
        </MenuItem>
      );
    });
  };

  const menuItems = getMenuItems();

  return (
    <Box>
      <FormControl sx={{ minWidth: 120 }} size="small">
        <Select
          sx={{ maxHeight: 48 }}
          id="grouped-select"
          autoWidth
          value={sourceId}
          onChange={(e) => {
            setSourceId(e.target.value);
          }}
        >
          {menuItems}
        </Select>
      </FormControl>
    </Box>
  );
};

export default TriggerSourceSelector;
