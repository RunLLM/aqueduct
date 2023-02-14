import { Box, Link, Typography } from '@mui/material';
import React, { useState } from 'react';

import { getPathPrefix } from '../../utils/getPathPrefix';
import {
  Integration,
  NotificationIntegrationConfig,
} from '../../utils/integrations';
import { NotificationLogLevel } from '../../utils/notifications';
import { Button } from '../primitives/Button.styles';
import NotificationLevelSelector from './NotificationLevelSelector';

export type NotificationConfigsMap = {
  [integrationId: string]: NotificationIntegrationConfig;
};

type Props = {
  notifications: Integration[];
  onSave: (updatedConfigs: NotificationConfigsMap) => void;
  isSaving: boolean;
};

function ConfigDifference(
  initialConfigs: NotificationConfigsMap,
  currentConfigs: NotificationConfigsMap
): NotificationConfigsMap {
  const results = {};
  Object.entries(currentConfigs).forEach(([k, v]) => {
    const initialV = initialConfigs[k];
    if (initialV.enabled !== v.enabled || initialV.level !== v.level) {
      results[k] = v;
    }
  });

  return results;
}

const AccountNotificationSettingsSelector: React.FC<Props> = ({
  onSave,
  notifications,
  isSaving,
}) => {
  const initialConfigs = Object.fromEntries(
    notifications.map((x) => [x.id, x.config as NotificationIntegrationConfig])
  );
  const [configs, setConfigs] =
    useState<NotificationConfigsMap>(initialConfigs);

  if (!notifications.length) {
    return (
      <Typography variant="body1">
        You do not have any notification configured. You can add new
        notifications from the{' '}
        <Link href={`${getPathPrefix()}/integrations`} target="_blank">
          integrations
        </Link>{' '}
        page.
      </Typography>
    );
  }

  const configDifference = ConfigDifference(initialConfigs, configs);
  const showSaveAndCancel = Object.keys(configDifference).length > 0;
  const onCancel = () => setConfigs(initialConfigs);

  return (
    <Box display="flex" flexDirection="column" alignItems="left">
      {notifications.map((n) => (
        <Box
          key={n.id}
          display="flex"
          flexDirection="column"
          alignItems="left"
          paddingY={1}
        >
          <NotificationLevelSelector
            level={configs[n.id].level as NotificationLogLevel}
            onSelectLevel={(level) =>
              setConfigs({
                ...configs,
                [n.id]: { enabled: configs[n.id].enabled, level: level },
              })
            }
            enabled={configs[n.id].enabled === 'true'}
            enableSelectorMessage={n.name}
            disabledMessage={`${n.name} notifications will not be configured for all workflows by default.`}
            onEnable={(enabled) =>
              setConfigs({
                ...configs,
                [n.id]: {
                  enabled: enabled ? 'true' : 'false',
                  level: configs[n.id].level,
                },
              })
            }
          />
        </Box>
      ))}
      {showSaveAndCancel && (
        <Box display="flex" flexDirection="row">
          <Button
            onClick={() => onSave(configDifference)}
            color="primary"
            disabled={isSaving}
          >
            Save
          </Button>
          <Button
            onClick={() => onCancel()}
            sx={{ marginLeft: 2 }}
            color="secondary"
          >
            Cancel
          </Button>
        </Box>
      )}
    </Box>
  );
};

export default AccountNotificationSettingsSelector;
