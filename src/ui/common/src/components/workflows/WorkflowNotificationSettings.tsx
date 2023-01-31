import { faPlus, faTrash } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Box, MenuItem, Select, Typography } from '@mui/material';
import React from 'react';
import { Integration } from 'src/utils/integrations';
import { NotificationLogLevel } from 'src/utils/notifications';
import { NotificationSettings } from 'src/utils/workflows';

import NotificationLevelSelector from '../notifications/NotificationLevelSelector';

type SelectedNotificationEntryProps = {
  remainingNotificationIntegrations: Integration[];
  selected: Integration;
  level: NotificationLogLevel | undefined;
  onSelect: (id: string, level: NotificationLogLevel | undefined) => void;
  onRemove: (id: string) => void;
};

export const SelectedNotificationEntry: React.FC<
  SelectedNotificationEntryProps
> = ({
  remainingNotificationIntegrations,
  selected,
  level,
  onSelect,
  onRemove,
}) => {
  return (
    <Box display="flex" flexDirection="column">
      <Box display="flex" flexDirection="row" alignItems="center">
        <Select autoWidth value={selected.id}>
          {[selected].concat(remainingNotificationIntegrations).map((x) => (
            <MenuItem
              key={selected.id + x.id}
              value={x.id}
              onClick={() => onSelect(x.id, level)}
            >
              <Typography>{x.name}</Typography>
            </MenuItem>
          ))}
        </Select>
        <Box ml={2}>
          <FontAwesomeIcon
            icon={faTrash}
            color="gray.600"
            onClick={() => onRemove(selected.id)}
          />
        </Box>
      </Box>
      <Box mt={1}>
        <NotificationLevelSelector
          level={level}
          onSelectLevel={(level) => onSelect(selected.id, level)}
        />
      </Box>
    </Box>
  );
};

type Props = {
  notificationIntegrations: Integration[];
  curSettings: NotificationSettings;
  onSelect: (id: string, level?: NotificationLogLevel) => void;
  onRemove: (id: string) => void;
};

const WorkflowNotificationSettings: React.FC<Props> = ({
  notificationIntegrations,
  curSettings,
  onSelect,
  onRemove,
}) => {
  const selectedIDs = Object.keys(curSettings);
  const remainingIntegrations = notificationIntegrations.filter(
    (x) => !selectedIDs.includes(x.id)
  );
  const integrationsByID: { [id: string]: Integration } = {};
  notificationIntegrations.forEach((x) => (integrationsByID[x.id] = x));

  const selectedEntries = Object.entries(curSettings).map(([id, level]) => (
    <Box key={id} mb={1}>
      <SelectedNotificationEntry
        remainingNotificationIntegrations={remainingIntegrations}
        selected={integrationsByID[id]}
        level={level}
        onSelect={onSelect}
        onRemove={onRemove}
      />
    </Box>
  ));

  return (
    <Box display="flex" flexDirection="column" alignContent="left">
      {selectedEntries}
      {remainingIntegrations.length > 0 && (
        <FontAwesomeIcon
          icon={faPlus}
          color="gray.800"
          onClick={() => onSelect(remainingIntegrations[0].id, undefined)}
        />
      )}
    </Box>
  );
};

export default WorkflowNotificationSettings;
