import { Divider } from '@mui/material';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React, { useState } from 'react';

import { SlackConfig } from '../../../utils/integrations';
import { NotificationLogLevel } from '../../../utils/notifications';
import CheckboxEntry from '../../notifications/CheckboxEntry';
import NotificationLevelSelector from '../../notifications/NotificationLevelSelector';
import { IntegrationTextInputField } from './IntegrationTextInputField';

// Placeholders are example values not filled for users, but
// may show up in textbox as hint if user don't fill the form field.
const Placeholders = {
  token: '*****',
  channel: 'my_channel',
  level: 'succeeded',
  enabled: true,
};

// Default fields are actual filled form values on 'create' dialog.
export const SlackDefaultsOnCreate = {
  token: '',
  channels_serialized: '',
  level: NotificationLogLevel.Success,
  enabled: 'false',
};

type Props = {
  onUpdateField: (field: keyof SlackConfig, value: string) => void;
  value?: SlackConfig;
};

export const SlackDialog: React.FC<Props> = ({ onUpdateField, value }) => {
  const [channels, setChannels] = useState(
    value?.channels_serialized
      ? (JSON.parse(value?.channels_serialized) as string[]).join(',')
      : ''
  );

  return (
    <Box sx={{ mt: 2 }}>
      <IntegrationTextInputField
        spellCheck={false}
        required={false}
        label="Bot Token *"
        description="The slack bot token. Please make sure this token has the permissions to send messages to channels you specified."
        placeholder={Placeholders.token}
        type="password"
        onChange={(event) => {
          onUpdateField('token', event.target.value);
        }}
        value={value?.token ?? null}
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Channels *"
        description="The channel(s) to send notifications. Use comma to separate different channels."
        placeholder={Placeholders.channel}
        onChange={(event) => {
          setChannels(event.target.value);
          const channelsList = event.target.value
            .split(',')
            .map((r) => r.trim());
          onUpdateField('channels_serialized', JSON.stringify(channelsList));
        }}
        value={channels ?? null}
      />

      <Divider sx={{ mt: 2 }} />

      <Box sx={{ mt: 2 }}>
        <CheckboxEntry
          checked={value?.enabled === 'true'}
          disabled={false}
          onChange={(checked) =>
            onUpdateField('enabled', checked ? 'true' : 'false')
          }
        >
          Enable this notification for all workflows.
        </CheckboxEntry>
        <Typography variant="body2" color="darkGray">
          Configure if we should apply this notification to all workflows unless
          separately specified in workflow settings.
        </Typography>
      </Box>

      {value?.enabled === 'true' && (
        <Box sx={{ mt: 2 }}>
          <Box sx={{ my: 1 }}>
            <Typography variant="body1" sx={{ fontWeight: 'bold' }}>
              Level
            </Typography>
            <Typography variant="body2" sx={{ color: 'darkGray' }}>
              The notification levels at which to send a slack notification.
              This applies to all workflows unless separately specified in
              workflow settings.
            </Typography>
          </Box>
          <NotificationLevelSelector
            level={value?.level as NotificationLogLevel}
            onSelectLevel={(level) => onUpdateField('level', level)}
            enabled={value?.enabled === 'true'}
          />
        </Box>
      )}
    </Box>
  );
};

export function isSlackConfigComplete(config: SlackConfig): boolean {
  if (config.enabled !== 'true' && config.enabled !== 'false') {
    return false;
  }

  if (config.enabled == 'true' && !config.level) {
    return false;
  }

  return !!config.channels_serialized && !!config.token;
}
