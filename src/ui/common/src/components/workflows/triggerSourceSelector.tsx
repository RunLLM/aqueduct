import Box from '@mui/material/Box';
import FormControl from '@mui/material/FormControl';
import MenuItem from '@mui/material/MenuItem';
import Select from '@mui/material/Select';
import Typography from '@mui/material/Typography';
import React from 'react';
import { useSelector } from 'react-redux';
import { ListWorkflowSummary } from 'src/utils/workflows';

import { RootState } from '../../stores/store';

type Props = {
  workflows: ListWorkflowSummary[];
};

export const TriggerSourceSelector: React.FC<Props> = ({ workflows }) => {
  const workflow = useSelector((state: RootState) => state.workflowReducer);

  const dag = workflow.selectedDag;

  const filteredWorkflows = workflows.filter(
    (workflow) => workflow.id === dag.metadata?.schedule.source_id
  );
  // TODO: ENG-2181 Add support for changing source trigger
  const selected = filteredWorkflows[0].id;

  const getMenuItems = () => {
    return filteredWorkflows.map((workflow) => {
      return (
        <MenuItem
          value={workflow.id}
          key={workflow.id}
          sx={{ backgroundColor: selected ? 'blueTint' : null }}
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

export default TriggerSourceSelector;
