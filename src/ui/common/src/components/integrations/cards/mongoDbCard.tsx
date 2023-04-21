import Box from '@mui/material/Box';
import React from 'react';

import { Integration, MongoDBConfig } from '../../../utils/integrations';
import {TruncatedText} from "./truncatedText";

type Props = {
  integration: Integration;
};

export const MongoDBCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as MongoDBConfig;
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      <TruncatedText variant="body2">
        <strong>URI: </strong>
        ********
      </TruncatedText>
      <TruncatedText variant="body2">
        <strong>Database: </strong>
        {config.database}
      </TruncatedText>
    </Box>
  );
};
