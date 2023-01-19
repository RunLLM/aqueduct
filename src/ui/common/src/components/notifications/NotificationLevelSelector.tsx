import { Checkbox, FormControlLabel, FormGroup } from '@mui/material';
import React from 'react';

import { NotificationLogLevel } from '../../utils/notifications';

type Props = {
  level?: NotificationLogLevel;
  onSelectLevel: (level?: NotificationLogLevel) => void;
};

const NotificationLevelSelector: React.FC<Props> = ({
  level,
  onSelectLevel,
}) => {
  return (
    <FormGroup>
      <FormControlLabel
        control={
          <Checkbox
            checked={[
              NotificationLogLevel.Success,
              NotificationLogLevel.Warning,
              NotificationLogLevel.Error,
            ].includes(level)}
            disabled={
              // disable if higher level has been selected
              [
                NotificationLogLevel.Success,
                NotificationLogLevel.Warning,
              ].includes(level)
            }
            onChange={(event) =>
              onSelectLevel(
                event.target.checked ? NotificationLogLevel.Error : undefined
              )
            }
          />
        }
        label="error"
      />
      <FormControlLabel
        control={
          <Checkbox
            checked={[
              NotificationLogLevel.Success,
              NotificationLogLevel.Warning,
            ].includes(level)}
            disabled={
              // disable if higher level has been selected
              [NotificationLogLevel.Success].includes(level)
            }
            onChange={(event) =>
              onSelectLevel(
                event.target.checked
                  ? NotificationLogLevel.Warning
                  : NotificationLogLevel.Error
              )
            }
          />
        }
        label="warning"
      />
      <FormControlLabel
        control={
          <Checkbox
            checked={level === NotificationLogLevel.Success}
            onChange={(event) =>
              onSelectLevel(
                event.target.checked
                  ? NotificationLogLevel.Success
                  : NotificationLogLevel.Warning
              )
            }
          />
        }
        label="succeeded"
      />
    </FormGroup>
  );
};

export default NotificationLevelSelector;
