import { faPlusSquare, faTrash } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Box, Link, MenuItem, Select, Typography } from '@mui/material';
import React, { useState } from 'react';

import { theme } from '../../styles/theme/theme';
import { getPathPrefix } from '../../utils/getPathPrefix';
import { Integration } from '../../utils/integrations';
import { NotificationLogLevel } from '../../utils/notifications';
import { NotificationSettingsMap } from '../../utils/workflows';
import CheckboxEntry from '../notifications/CheckboxEntry';
import NotificationLevelSelector from '../notifications/NotificationLevelSelector';

type SelectedNotificationEntryProps = {
  remainingNotificationIntegrations: Integration[];
  selected: Integration;
  level: NotificationLogLevel | undefined;
  onSelect: (
    id: string,
    level: NotificationLogLevel | undefined,
    replacingID?: string
  ) => void;
  onRemove: (id: string) => void;
};

type Props = {
  notificationIntegrations: Integration[];
  curSettingsMap: NotificationSettingsMap;
  onSelect: (
    id: string,
    level?: NotificationLogLevel,
    replacingID?: string
  ) => void;
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
        <Select autoWidth sx={{ height: 36 }} value={selected.id}>
          {[selected]
            .concat(remainingNotificationIntegrations) // show current + remaining as options
            .sort((x, y) => (x.name > y.name ? 1 : -1)) // sort to ensure items are stable
            .map((x) => (
              <MenuItem
                key={x.id}
                value={x.id}
                onClick={() => onSelect(x.id, level, selected.id)}
              >
                <Typography>{x.name}</Typography>
              </MenuItem>
            ))}
        </Select>
        <Box ml={2}>
          <FontAwesomeIcon
            icon={faTrash}
            color={theme.palette.gray[700]}
            style={{ cursor: 'pointer' }}
            onClick={() => onRemove(selected.id)}
          />
        </Box>
      </Box>
      <Box mt={1}>
        <NotificationLevelSelector
          level={level}
          onSelectLevel={(level) => onSelect(selected.id, level)}
          disabled={false}
        />
      </Box>
    </Box>
  );
};

const WorkflowNotificationSettings: React.FC<Props> = ({
  notificationIntegrations,
  curSettingsMap,
  onSelect,
  onRemove,
}) => {
  const selectedIDs = Object.keys(curSettingsMap);
  const [usingDefault, setUsingDefault] = useState(selectedIDs.length === 0);
  const remainingIntegrations = notificationIntegrations.filter(
    (x) => !selectedIDs.includes(x.id)
  );
  const integrationsByID: { [id: string]: Integration } = {};
  notificationIntegrations.forEach((x) => (integrationsByID[x.id] = x));

  const selectedEntries = Object.entries(curSettingsMap).map(([id, level]) => (
    <Box key={id} mt={1}>
      <SelectedNotificationEntry
        remainingNotificationIntegrations={remainingIntegrations}
        selected={integrationsByID[id]}
        level={level}
        onSelect={onSelect}
        onRemove={onRemove}
      />
    </Box>
  ));

  const usingDefaultCheckbox = (
    <CheckboxEntry
      checked={usingDefault}
      onChange={(checked) => setUsingDefault(checked)}
    >
      Use default notification settings. See{' '}
      <Link
        underline="none"
        href={`${getPathPrefix()}/account`}
        target="_blank"
      >
        settings
      </Link>{' '}
      for more details.
    </CheckboxEntry>
  );

  return (
    <Box display="flex" flexDirection="column" alignContent="left">
      {<Box marginY={1}>{usingDefaultCheckbox}</Box>}
      {!usingDefault && selectedEntries}
      {!usingDefault && remainingIntegrations.length > 0 && (
        <Box mt={selectedEntries.length > 0 ? 2 : 1}>
          <FontAwesomeIcon
            icon={faPlusSquare}
            color={theme.palette.gray[700]}
            width="24px"
            fontSize="24px"
            style={{ cursor: 'pointer' }}
            onClick={() => onSelect(remainingIntegrations[0].id, undefined)}
          />
        </Box>
      )}
    </Box>
  );
};

export default WorkflowNotificationSettings;
