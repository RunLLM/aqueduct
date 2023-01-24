import { Box, Checkbox, Typography } from '@mui/material';
import React from 'react';

import { theme } from '../../styles/theme/theme';
import { NotificationLogLevel } from '../../utils/notifications';

type Props = {
  level?: NotificationLogLevel;
  onSelectLevel: (level?: NotificationLogLevel) => void;
};

const NotificationLevelSelector: React.FC<Props> = ({
  level,
  onSelectLevel,
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

  // disable if higher level has been checked
  const errorDisabled = warningChecked;
  const warningDisabled = successChecked;
  return (
    <Box display="flex" flexDirection="column" alignContent="left">
      <Box display="flex" flexDirection="row" alignContent="center">
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
          ERROR
        </Typography>
      </Box>
      <Box
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
          WARNING
        </Typography>
      </Box>
      <Box
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
          SUCCESS
        </Typography>
      </Box>
    </Box>
  );
};

export default NotificationLevelSelector;
