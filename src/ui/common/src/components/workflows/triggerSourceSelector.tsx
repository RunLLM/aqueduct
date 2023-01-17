import Box from '@mui/material/Box';
import FormControl from '@mui/material/FormControl';
import MenuItem from '@mui/material/MenuItem';
import Select from '@mui/material/Select';
import Typography from '@mui/material/Typography';
import React, { useEffect, useState } from 'react';
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
  const [selected, setSelected] = useState<ListWorkflowSummary>(
    workflows.find((workflow) => workflow.id === sourceId)
  );

  useEffect(() => {
    if (!selected) {
      return;
    }

    setSourceId(selected.id);
  }, [selected, setSourceId]);

  const getMenuItems = () => {
    return workflows.map((workflow) => {
      return (
        //@ts-ignore
        <MenuItem
          key={workflow.id}
          value={workflow}
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
          value={selected}
          onChange={(e) => {
            setSelected(e.target.value as ListWorkflowSummary);
          }}
        >
          {menuItems}
        </Select>
      </FormControl>
    </Box>
  );
};

export default TriggerSourceSelector;
