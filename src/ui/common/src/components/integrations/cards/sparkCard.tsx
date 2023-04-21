import Box from '@mui/material/Box';
import React from 'react';
import { TruncatedText } from './truncatedText';
import { Integration, SparkConfig } from '../../../utils/integrations';

type SparkCardProps = {
  integration: Integration;
};

export const SparkCard: React.FC<SparkCardProps> = ({ integration }) => {
  const config = integration.config as SparkConfig;
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      <TruncatedText variant="body2">
        <strong>Livy Server URL: </strong>
        {config.livy_server_url}
      </TruncatedText>
    </Box>
  );
};

export default SparkCard;
