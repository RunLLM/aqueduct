import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React, { useState } from 'react';
import { NotificationLogLevel } from 'src';

import { EmailConfig } from '../../../utils/integrations';
import NotificationLevelSelector from '../../notifications/NotificationLevelSelector';
import { IntegrationTextInputField } from './IntegrationTextInputField';

const Placeholders = {
  host: 'smtp.myprovider.com',
  port: '',
  user: 'mysender@myprovider.com',
  password: '******',
  reciever: 'myreciever@myprovider.com',
  level: 'succeeded',
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
            Level *
          </Typography>
          <Typography variant="body2" sx={{ color: 'darkGray' }}>
            The notification levels at which to send an email notification. This
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
