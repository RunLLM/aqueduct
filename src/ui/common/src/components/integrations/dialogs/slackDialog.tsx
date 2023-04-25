import { Divider } from '@mui/material';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React, { useState } from 'react';
import { useFormContext } from 'react-hook-form';

import {
  IntegrationDialogProps,
  SlackConfig,
} from '../../../utils/integrations';
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

export const SlackDialog: React.FC<IntegrationDialogProps> = ({
  editMode = false,
}) => {
  const [selectedLevel, setSelectedLevel] = useState(
    SlackDefaultsOnCreate.level
  );
  const [notificationsEnabled, setNotificationsEnabled] = useState(
    SlackDefaultsOnCreate.enabled
  );

  const { register, setValue, getValues } = useFormContext();
  register('level', { value: SlackDefaultsOnCreate.level });
  register('enabled', { value: SlackDefaultsOnCreate.enabled });
  register('channels_serialized', {
    value: SlackDefaultsOnCreate.channels_serialized,
  });
  const channels_serialized = getValues('channels_serialized');
  register('channels', {
    value: channels_serialized
      ? (JSON.parse(channels_serialized) as string[]).join(',')
      : '',
  });

  return (
    <Box sx={{ mt: 2 }}>
      <IntegrationTextInputField
        name="token"
        spellCheck={false}
        required={false}
        label="Bot Token *"
        description="The slack bot token. Please make sure this token has the permissions to send messages to channels you specified."
        placeholder={Placeholders.token}
        type="password"
        onChange={(event) => {
          setValue('token', event.target.value);
        }}
      />

      <IntegrationTextInputField
        name="channels"
        spellCheck={false}
        required={true}
        label="Channels *"
        description="The channel(s) to send notifications. Use comma to separate different channels."
        placeholder={Placeholders.channel}
        onChange={(event) => {
          setValue('channels', event.target.value);
          const channelsList = event.target.value
            .split(',')
            .map((r) => r.trim());

          const serializedChannels = JSON.stringify(channelsList);
          setValue('channels_serialized', serializedChannels);
        }}
      />

      <Divider sx={{ mt: 2 }} />

      <Box sx={{ mt: 2 }}>
        <CheckboxEntry
          checked={notificationsEnabled === 'true'}
          disabled={false}
          onChange={(checked) => {
            setNotificationsEnabled(checked ? 'true' : 'false');
            setValue('enabled', checked ? 'true' : 'false');
          }}
        >
          Enable this notification for all workflows.
        </CheckboxEntry>
        <Typography variant="body2" color="darkGray">
          Configure if we should apply this notification to all workflows unless
          separately specified in workflow settings.
        </Typography>
      </Box>

      {notificationsEnabled === 'true' && (
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
            level={selectedLevel as NotificationLogLevel}
            onSelectLevel={(level) => {
              setSelectedLevel(level);
              setValue('level', level);
            }}
            enabled={notificationsEnabled === 'true'}
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
