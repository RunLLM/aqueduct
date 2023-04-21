import Box from '@mui/material/Box';
import React from 'react';

import { Integration } from '../../../utils/integrations';
import { S3Config } from '../../../utils/workflows';
import {TruncatedText} from "./truncatedText";


type Props = {
  integration: Integration;
};

export const S3Card: React.FC<Props> = ({ integration }) => {
  const config = integration.config as S3Config;

  return (
    <Box>
      <TruncatedText variant="body2">
        <strong>Bucket: </strong>
        {config.bucket}
      </TruncatedText>
      {config.root_dir?.length > 0 && (
        <TruncatedText variant="body2">
          <strong>Root Directory: </strong>
          {config.root_dir}
        </TruncatedText>
      )}
      <TruncatedText variant="body2">
        <strong>Region: </strong>
        {config.region}
      </TruncatedText>
    </Box>
  );
};
