import Box from '@mui/material/Box';
import React from 'react';

import { Integration } from '../../../utils/integrations';
import { S3Config } from '../../../utils/workflows';
import { CardTextEntry } from './text';

type Props = {
  integration: Integration;
};

// This should be set to the minimum width required to display the longest category name on the card.
const categoryWidth = '100px';

export const S3Card: React.FC<Props> = ({ integration }) => {
  const config = integration.config as S3Config;

  return (
    <Box>
      <CardTextEntry
        category="Bucket: "
        value={config.bucket}
        categoryWidth={categoryWidth}
      />

      {config.root_dir?.length > 0 && (
        <CardTextEntry
          category="Root Directory: "
          value={config.root_dir}
          categoryWidth={categoryWidth}
        />
      )}

      <CardTextEntry
        category="Region: "
        value={config.region}
        categoryWidth={categoryWidth}
      />
    </Box>
  );
};
