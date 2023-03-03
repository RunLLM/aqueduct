import { Box, Typography } from '@mui/material';
import React from 'react';
import { ServiceLogos } from '../../../../utils/integrations';

export interface EngineItemProps {
  engine: string;
}

export const EngineItem: React.FC<EngineItemProps> = ({
  // The expectation is that we get the internal representation of the engine name, 
  // which is all lowercase.
  engine,
}) => {
  const engineName = engine[0].toUpperCase() + engine.substring(1);
  const iconUrl = ServiceLogos[engineName];

  return (
    <Box display="flex" alignItems="left" justifyContent="left">
      <img
        src={iconUrl}
        style={{ marginTop: '4px', marginRight: '8px' }}
        width="16px"
        height="16px"
      />
      <Typography variant="body1">{engineName}</Typography>
    </Box>
  );
};

export default EngineItem;
