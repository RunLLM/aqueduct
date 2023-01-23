import {
  faCircleCheck,
  faCircleExclamation,
  faMinus,
  faTriangleExclamation,
} from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import Box from '@mui/material/Box';
import React from 'react';

import { theme } from '../../styles/theme/theme';

interface CheckTableItemProps {
  checkValue: string;
}

export const CheckTableItem: React.FC<CheckTableItemProps> = ({
  checkValue,
}) => {
  let iconColor = theme.palette.black;
  let checkIcon = faMinus;

  switch (checkValue.toLowerCase()) {
    case 'true': {
      checkIcon = faCircleCheck;
      iconColor = theme.palette.Success;
      break;
    }
    case 'false': {
      checkIcon = faCircleExclamation;
      iconColor = theme.palette.red['500'];
      break;
    }
    case 'warning': {
      checkIcon = faTriangleExclamation;
      iconColor = theme.palette.orange['500'];
      break;
    }
    case 'none': {
      checkIcon = faMinus;
      iconColor = theme.palette.black;
      break;
    }
    default: {
      // None of the icon cases met, just fall through and render table value.
      return <>{checkValue}</>;
    }
  }

  return (
    <Box sx={{ fontSize: '16px', color: iconColor }}>
      <FontAwesomeIcon icon={checkIcon} />
    </Box>
  );
};

export default CheckTableItem;
