import { Box, Checkbox, Typography } from '@mui/material';
import React from 'react';

import { theme } from '../../styles/theme/theme';
import { NotificationLogLevel } from '../../utils/notifications';

type Props = {
  level?: NotificationLogLevel;
  onSelectLevel: (level?: NotificationLogLevel) => void;
  // if set, we will show an additional option to allow disabling the notification
  // using the given message.
  disabled: boolean;
  disabledMessage?: string;
  disableSelectorMessage?: string;
  onDisable?: (disabled: boolean) => void;
};

const NotificationLevelSelector: React.FC<Props> = ({
  level,
  onSelectLevel,
  disabled,
  disabledMessage,
  disableSelectorMessage,
  onDisable,
}) => {
  // Overrides default checkbox behavior.
  // The `padding` is particularly important as the checkbox has a default padding of 9.
  const checkboxStyle = {
    padding: 0,
    '&.Mui-checked': { color: theme.palette.blue[700] },
    '&.Mui-disabled': { color: theme.palette.gray[700] },
  };

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

  const levelSelectorLeftMargin = showDisableOption ? 2 : 0;

  return (
    <Box display="flex" flexDirection="column" alignContent="left">
      {showDisableOption && (
        <Box
          display="flex"
          flexDirection="row"
          alignContent="center"
          marginTop={1}
        >
          <Checkbox
            checked={disabled}
            onChange={(event) => {
              if (!event.target.checked && !level) {
                onSelectLevel(NotificationLogLevel.Success);
              }

              if (!!onDisable) {
                onDisable(event.target.checked);
              }
            }}
            sx={checkboxStyle}
          />
          <Typography
            variant="body1"
            color="black"
            marginLeft={1}
            alignContent="center"
          >
            {disableSelectorMessage}
          </Typography>
        </Box>
      )}
      {disabled && !!disabledMessage && (
        <Typography variant="body2" color="gray.700">
          {disabledMessage}
        </Typography>
      )}
      {showLevelOptions && (
        <Box
          marginLeft={levelSelectorLeftMargin}
          display="flex"
          flexDirection="row"
          alignContent="center"
        >
          <Checkbox
            checked={errorChecked}
            disabled={errorDisabled}
            onChange={(event) =>
              onSelectLevel(
                event.target.checked ? NotificationLogLevel.Error : undefined
              )
            }
            sx={checkboxStyle}
          />
          <Typography
            variant="body1"
            color={errorDisabled ? 'gray.700' : 'black'}
            marginLeft={1}
          >
            Error
          </Typography>
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
          <Checkbox
            checked={warningChecked}
            disabled={warningDisabled}
            onChange={(event) =>
              onSelectLevel(
                event.target.checked
                  ? NotificationLogLevel.Warning
                  : NotificationLogLevel.Error
              )
            }
            sx={checkboxStyle}
          />
          <Typography
            variant="body1"
            color={warningDisabled ? 'gray.700' : 'black'}
            marginLeft={1}
          >
            Warning
          </Typography>
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
          <Checkbox
            checked={successChecked}
            onChange={(event) =>
              onSelectLevel(
                event.target.checked
                  ? NotificationLogLevel.Success
                  : NotificationLogLevel.Warning
              )
            }
            sx={checkboxStyle}
          />
          <Typography
            variant="body1"
            color="black"
            marginLeft={1}
            alignContent="center"
          >
            Success
          </Typography>
        </Box>
      )}
    </Box>
  );
};

export default NotificationLevelSelector;
