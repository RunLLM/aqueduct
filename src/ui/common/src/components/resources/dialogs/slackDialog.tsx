import { Divider } from '@mui/material';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React, { useState } from 'react';
import { useFormContext } from 'react-hook-form';
import * as Yup from 'yup';

import { NotificationLogLevel } from '../../../utils/notifications';
import { ResourceDialogProps, SlackConfig } from '../../../utils/resources';
import CheckboxEntry from '../../notifications/CheckboxEntry';
import NotificationLevelSelector from '../../notifications/NotificationLevelSelector';
import { ResourceTextInputField } from './ResourceTextInputField';
import { requiredAtCreate } from './schema';

// Placeholders are example values not filled for users, but
// may show up in textbox as hint if user don't fill the form field.
const Placeholders = {
  token: '*****',
  channel: 'my_channel',
  level: 'succeeded',
  enabled: 'false',
};

// Default fields are actual filled form values on 'create' dialog.
export const SlackDefaultsOnCreate: SlackConfig = {
  token: '',
  channels_serialized: '',
  level: NotificationLogLevel.Success,
  enabled: 'false',
};

export const SlackDialog: React.FC<ResourceDialogProps<SlackConfig>> = ({
  resourceToEdit,
}) => {
  const initialLevel = resourceToEdit?.level ?? SlackDefaultsOnCreate.level;
  const initialEnabled =
    resourceToEdit?.enabled ?? SlackDefaultsOnCreate.enabled;
  const [selectedLevel, setSelectedLevel] = useState(initialLevel);

  const [notificationsEnabled, setNotificationsEnabled] =
    useState(initialEnabled);

  const { register, setValue } = useFormContext();

  if (resourceToEdit) {
    Object.entries(resourceToEdit).forEach(([k, v]) => {
      register(k, { value: v });
    });
  } else {
    register('enabled', { value: SlackDefaultsOnCreate.enabled });
    register('level', { value: SlackDefaultsOnCreate.level });
    register('channels_serialized', {
      value: SlackDefaultsOnCreate.channels_serialized,
    });
  }

  return (
    <Box sx={{ mt: 2 }}>
      <ResourceTextInputField
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

      <ResourceTextInputField
        name="channels"
        spellCheck={false}
        required={true}
        label="Channels *"
        description="The channel(s) to send notifications. Use comma to separate different channels."
        placeholder={Placeholders.channel}
        onChange={(event) => {
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

export function getSlackValidationSchema(editMode: boolean) {
  return Yup.object().shape({
    token: requiredAtCreate(Yup.string(), editMode, 'Please enter a token'),
    channels_serialized: Yup.string().required(
      'Please enter at least one channel name'
    ),
    level: Yup.string().required('Please select a notification level'),
    enabled: Yup.string().required(),
  });
}
