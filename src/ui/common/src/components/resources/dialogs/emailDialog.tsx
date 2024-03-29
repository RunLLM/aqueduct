import { Divider } from '@mui/material';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React, { useState } from 'react';
import { useFormContext } from 'react-hook-form';
import * as Yup from 'yup';

import { NotificationLogLevel } from '../../../utils/notifications';
import { EmailConfig, ResourceDialogProps } from '../../../utils/resources';
import CheckboxEntry from '../../notifications/CheckboxEntry';
import NotificationLevelSelector from '../../notifications/NotificationLevelSelector';
import { ResourceTextInputField } from './ResourceTextInputField';
import { requiredAtCreate } from './schema';

// Placeholders are example values not filled for users, but
// may show up in textbox as hint if user don't fill the form field.
const Placeholders = {
  host: 'smtp.myprovider.com',
  port: '',
  user: 'mysender@myprovider.com',
  password: '******',
  reciever: 'myreciever@myprovider.com',
  level: 'succeeded',
  enabled: 'false',
};

// Default fields are actual filled form values on 'create' dialog.
export const EmailDefaultsOnCreate = {
  host: '',
  port: '',
  user: '',
  password: '',
  targets_serialized: '',
  level: NotificationLogLevel.Success,
  enabled: 'false',
};

export const EmailDialog: React.FC<ResourceDialogProps<EmailConfig>> = ({
  resourceToEdit,
}) => {
  const initialLevel = resourceToEdit?.level ?? EmailDefaultsOnCreate.level;
  const initialEnabled =
    resourceToEdit?.enabled ?? EmailDefaultsOnCreate.enabled;
  const [selectedLevel, setSelectedLevel] = useState(initialLevel);
  const [notificationsEnabled, setNotificationsEnabled] =
    useState(initialEnabled);
  const { register, setValue } = useFormContext();

  if (resourceToEdit) {
    Object.entries(resourceToEdit).forEach(([k, v]) => {
      register(k, { value: v });
    });
  } else {
    register('enabled', { value: EmailDefaultsOnCreate.enabled });
    register('level', { value: EmailDefaultsOnCreate.level });
    register('targets_serialized', {
      value: EmailDefaultsOnCreate.targets_serialized,
    });
  }

  return (
    <Box sx={{ mt: 2 }}>
      <ResourceTextInputField
        name="host"
        spellCheck={false}
        required={true}
        label="Host *"
        description="The hostname address of the email SMTP server."
        placeholder={Placeholders.host}
        onChange={(event) => setValue('host', event.target.value)}
      />

      <ResourceTextInputField
        name="port"
        spellCheck={false}
        required={true}
        label="Port *"
        description="The port number of the email SMTP server."
        placeholder={Placeholders.port}
        onChange={(event) => setValue('port', event.target.value)}
      />

      <ResourceTextInputField
        name="user"
        spellCheck={false}
        required={true}
        label="Sender Address *"
        description="The email address of the sender."
        placeholder={Placeholders.user}
        onChange={(event) => setValue('user', event.target.value)}
      />

      <ResourceTextInputField
        name="password"
        spellCheck={false}
        required={false}
        label="Sender Password *"
        description="The password corresponding to the above email address."
        placeholder={Placeholders.password}
        type="password"
        onChange={(event) => {
          setValue('password', event.target.value);
        }}
      />

      <ResourceTextInputField
        name="receivers"
        spellCheck={false}
        required={true}
        label="Receiver Address *"
        description="The email address(es) of the receiver(s). Use comma to separate different addresses."
        placeholder={Placeholders.reciever}
        onChange={(event) => {
          const receiversList = event.target.value
            .split(',')
            .map((r) => r.trim());
          setValue('targets_serialized', JSON.stringify(receiversList));
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
              The notification levels at which to send an email notification.
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

export function getEmailValidationSchema(editMode: boolean) {
  return Yup.object().shape({
    host: Yup.string().required('Please enter a host'),
    port: Yup.string().required('Please enter a port'),
    user: Yup.string().required('Please enter a sender address'),
    password: requiredAtCreate(
      Yup.string(),
      editMode,
      'Please enter a sender password'
    ),
    targets_serialized: Yup.string().required(
      'Please enter at least one receiver'
    ),
    level: Yup.string().required('Please select a notification level'),
    enabled: Yup.string().required(),
  });
}
