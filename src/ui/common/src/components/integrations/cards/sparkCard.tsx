import { Typography } from '@mui/material';
import Box from '@mui/material/Box';
import React from 'react';

import { Integration, SparkConfig } from '../../../utils/integrations';

type SparkCardProps = {
    integration: Integration;
};

export const SparkCard: React.FC<SparkCardProps> = ({ integration }) => {
    const config = integration.config as SparkConfig;
    return (
        <Box sx={{ display: 'flex', flexDirection: 'column' }}>
            <Typography variant="body2">
                <strong>App Name: </strong>
                {config.app_name}
            </Typography>
            <Typography variant="body2">
                <strong>Driver Host: </strong>
                {config.driver_host}
            </Typography>
            <Typography variant="body2">
                <strong>Master: </strong>
                {config.master}
            </Typography>
        </Box>
    );
};
