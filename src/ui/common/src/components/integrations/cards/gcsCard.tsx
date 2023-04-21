import Box from '@mui/material/Box';
import React from 'react';

import { GCSConfig, Integration } from '../../../utils/integrations';
import {TruncatedText} from "./truncatedText";

type Props = {
  integration: Integration;
};

export const GCSCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as GCSConfig;

  return (
    <Box>
      <TruncatedText variant="body2">
        <strong>Bucket: </strong>
        {config.bucket}
      </TruncatedText>
    </Box>
  );
};
