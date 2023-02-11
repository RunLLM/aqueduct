import { Box } from '@mui/material';
import React from 'react';

import { NotificationLogLevel } from '../../utils/notifications';
import CheckboxEntry from './CheckboxEntry';

type Props = {
  level?: NotificationLogLevel;
  onSelectLevel: (level?: NotificationLogLevel) => void;
  // if set, we will show an additional option to allow disabling the notification
  // using the given message.
  disabled: boolean;
  disableSelectorMessage?: string;
  onDisable?: (disabled: boolean) => void;
};

const NotificationLevelSelector: React.FC<Props> = ({
  level,
  onSelectLevel,
  disabled,
  disableSelectorMessage,
  onDisable,
}) => {
  const errorChecked = [
    NotificationLogLevel.Success,
    NotificationLogLevel.Warning,
    NotificationLogLevel.Error,
  ].includes(level);
  const warningChecked = [
    NotificationLogLevel.Success,
    NotificationLogLevel.Warning,
  ].includes(level);
  const successChecked = level === NotificationLogLevel.Success;

  const showDisableOption = !!disableSelectorMessage;
  // show level if either:
  // * showing disable options and the option is unchecked
  // * not showing disable options
  const showLevelOptions =
    (!disabled && showDisableOption) || !showDisableOption;

  // disable if higher level has been checked
  const errorDisabled = warningChecked;
  const warningDisabled = successChecked;

  return (
    <Box display="flex" flexDirection="column" alignContent="left">
      {showLevelOptions && (
        <Box marginTop={1}>
          <CheckboxEntry
            checked={errorChecked}
            disabled={errorDisabled}
            onChange={(checked) =>
              onSelectLevel(checked ? NotificationLogLevel.Error : undefined)
            }
          >
            Error
          </CheckboxEntry>
        </Box>
      )}
      {showLevelOptions && (
        <Box marginTop={1}>
          <CheckboxEntry
            checked={warningChecked}
            disabled={warningDisabled}
            onChange={(checked) =>
              onSelectLevel(
                checked
                  ? NotificationLogLevel.Warning
                  : NotificationLogLevel.Error
              )
            }
          >
            Warning
          </CheckboxEntry>
        </Box>
      )}
      {showLevelOptions && (
        <Box
          display="flex"
          flexDirection="row"
          alignContent="center"
          marginTop={1}
        >
          <CheckboxEntry
            checked={successChecked}
            onChange={(checked) =>
              onSelectLevel(
                checked
                  ? NotificationLogLevel.Success
                  : NotificationLogLevel.Warning
              )
            }
          >
            Success
          </CheckboxEntry>
        </Box>
      )}
      {showDisableOption && (
        <Box marginTop={1}>
          <CheckboxEntry
            checked={disabled}
            onChange={(checked) => {
              if (!checked && !level) {
                onSelectLevel(NotificationLogLevel.Success);
              }

              if (!!onDisable) {
                onDisable(checked);
              }
            }}
          >
            {disableSelectorMessage}
          </CheckboxEntry>
        </Box>
      )}
    </Box>
  );
};

export default NotificationLevelSelector;
