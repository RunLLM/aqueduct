import { Box, Typography } from '@mui/material';
import React from 'react';

import { NotificationLogLevel } from '../../utils/notifications';
import CheckboxEntry from './CheckboxEntry';

type Props = {
  level?: NotificationLogLevel;
  onSelectLevel: (level?: NotificationLogLevel) => void;
  // if set, we will show an additional option to allow enabling the notification
  // using the given message.
  enabled: boolean;
  disabledMessage?: string;
  enableSelectorMessage?: string;
  onEnable?: (enabled: boolean) => void;
};

const NotificationLevelSelector: React.FC<Props> = ({
  level,
  onSelectLevel,
  enabled,
  disabledMessage,
  enableSelectorMessage,
  onEnable,
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

  const showEnableOption = !!enableSelectorMessage;
  // show level if either:
  // * showing enable options and the option is checked
  // * not showing enable options
  const showLevelOptions = (enabled && showEnableOption) || !showEnableOption;

  // disable if higher level has been checked
  const errorDisabled = warningChecked;
  const warningDisabled = successChecked;

  const levelSelectorLeftMargin = showEnableOption ? '30px' : undefined;

  return (
    <Box display="flex" flexDirection="column" alignContent="left">
      {showEnableOption && (
        <Box
          display="flex"
          flexDirection="row"
          alignContent="center"
          marginTop={1}
        >
          <CheckboxEntry
            checked={enabled}
            onChange={(checked) => {
              if (!checked && !level) {
                onSelectLevel(NotificationLogLevel.Success);
              }

              if (!!onEnable) {
                onEnable(checked);
              }
            }}
          >
            {enableSelectorMessage}
          </CheckboxEntry>
        </Box>
      )}
      {!enabled && !!disabledMessage && (
        <Typography variant="body2" color="gray.700">
          {disabledMessage}
        </Typography>
      )}
      {showLevelOptions && (
        <Box
          marginLeft={levelSelectorLeftMargin}
          marginTop={1}
          display="flex"
          flexDirection="row"
          alignContent="center"
        >
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
        <Box
          marginLeft={levelSelectorLeftMargin}
          display="flex"
          flexDirection="row"
          alignContent="center"
          marginTop={1}
        >
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
          marginLeft={levelSelectorLeftMargin}
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
    </Box>
  );
};

export default NotificationLevelSelector;
