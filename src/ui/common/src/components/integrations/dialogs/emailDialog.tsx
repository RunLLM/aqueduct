import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React, { useState } from 'react';

import { EmailConfig } from '../../../utils/integrations';
import { NotificationLogLevel } from '../../../utils/notifications';
import NotificationLevelSelector from '../../notifications/NotificationLevelSelector';
import { IntegrationTextInputField } from './IntegrationTextInputField';

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

type Props = {
  onUpdateField: (field: keyof EmailConfig, value: string) => void;
  value?: EmailConfig;
};

export const EmailDialog: React.FC<Props> = ({ onUpdateField, value }) => {
  const [receivers, setReceivers] = useState(
    value?.targets_serialized
      ? (JSON.parse(value?.targets_serialized) as string[]).join(',')
      : ''
  );

  return (
    <Box sx={{ mt: 2 }}>
      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Host *"
        description="The hostname address of the email SMTP server."
        placeholder={Placeholders.host}
        onChange={(event) => onUpdateField('host', event.target.value)}
        value={value?.host ?? null}
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Port *"
        description="The port number of the email SMTP server."
        placeholder={Placeholders.port}
        onChange={(event) => onUpdateField('port', event.target.value)}
        value={value?.port ?? null}
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Sender Address *"
        description="The email address of the sender."
        placeholder={Placeholders.user}
        onChange={(event) => onUpdateField('user', event.target.value)}
        value={value?.user ?? null}
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={false}
        label="Sender Password *"
        description="The password corresponding to the above email address."
        placeholder={Placeholders.password}
        type="password"
        onChange={(event) => {
          if (!!event.target.value) {
            onUpdateField('password', event.target.value);
          }
        }}
        value={value?.password ?? null}
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Receiver Address *"
        description="The email address(es) of the receiver(s). Use comma to separate different addresses."
        placeholder={Placeholders.reciever}
        onChange={(event) => {
          setReceivers(event.target.value);
          const receiversList = event.target.value
            .split(',')
            .map((r) => r.trim());
          onUpdateField('targets_serialized', JSON.stringify(receiversList));
        }}
        value={receivers ?? null}
      />

      <Box sx={{ mt: 2 }}>
        <Box sx={{ my: 1 }}>
          <Typography variant="body1" sx={{ fontWeight: 'bold' }}>
            How to apply to workflows *
          </Typography>
          <Typography variant="body2" sx={{ color: 'darkGray' }}>
            Configure if this email notification applies to all workflows. If
            enabled, choose level(s) on which to recieve notifications. You can
            override this setting in each workflow&apos;s settings page.
          </Typography>
        </Box>
        <NotificationLevelSelector
          level={value?.level as NotificationLogLevel}
          onSelectLevel={(level) => onUpdateField('level', level)}
          disableSelectorMessage="Do not apply this notification to all workflows."
          disabled={value?.enabled === 'false'}
          onDisable={(disabled) =>
            onUpdateField('enabled', disabled ? 'false' : 'true')
          }
        />
      </Box>
    </Box>
  );
};

export function isEmailConfigComplete(config: EmailConfig): boolean {
  if (config.enabled !== 'true' && config.enabled !== 'false') {
    return false;
  }

  if (config.enabled == 'true' && !config.level) {
    return false;
  }

  return (
    !!config.host &&
    !!config.port &&
    !!config.password &&
    !!config.targets_serialized &&
    !!config.user
  );
}
