import { Typography } from '@mui/material';
 import Box from '@mui/material/Box';
 import React from 'react';

 import { SparkConfig, Integration } from '../../../utils/integrations';

 type SparkCardProps = {
   integration: Integration;
 };

 export const SparkCard: React.FC<SparkCardProps> = ({
   integration,
 }) => {
   const config = integration.config as SparkConfig;
   return (
     <Box sx={{ display: 'flex', flexDirection: 'column' }}>
       <Typography variant="body2">
         <strong>Livy Server URL: </strong>
         {config.livy_server_url}
       </Typography>
     </Box>
   );
 };

 export default SparkCard;