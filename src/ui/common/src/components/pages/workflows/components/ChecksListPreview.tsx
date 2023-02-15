import { Box } from '@mui/material';
import React from 'react';

import { CheckPreview, getCheckStatusIcon } from './CheckItem';

interface ChecksListPreviewProps {
  checks: CheckPreview[];
}

export const ChecksListPreview: React.FC<ChecksListPreviewProps> = ({
  checks,
}) => {
  const checkIcons: JSX.Element[] = checks.map((check) =>
    getCheckStatusIcon(check)
  );

  return (
    <Box
      sx={{
        display: 'flex',
        width: '100%',
        overflow: 'hidden',
        whiteSpace: 'nowrap',
        textOverflow: 'ellipsis',
      }}
    >
      {checkIcons}
    </Box>
  );
};

export default ChecksListPreview;
