import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React, { useState } from 'react';
import { NotificationLogLevel } from 'src';

import { SlackConfig } from '../../../utils/integrations';
import NotificationLevelSelector from '../../notifications/NotificationLevelSelector';
import { IntegrationTextInputField } from './IntegrationTextInputField';

const Placeholders = {
  token: '*****',
  channel: 'my_channel',
  level: 'succeeded',
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
          if (!!event.target.value) {
            onUpdateField('token', event.target.value);
          }
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

      <Box sx={{ mt: 2 }}>
        <Box sx={{ my: 1 }}>
          <Typography variant="body1" sx={{ fontWeight: 'bold' }}>
            Level *
          </Typography>
          <Typography variant="body2" sx={{ color: 'darkGray' }}>
            The notification levels at which to send a slack notification. This
            applies to all workflows unless separately specified in workflow
            settings.
          </Typography>
        </Box>
        <NotificationLevelSelector
          level={value?.level as NotificationLogLevel}
          onSelectLevel={(level) => onUpdateField('level', level)}
        />
      </Box>
    </Box>
  );
};
