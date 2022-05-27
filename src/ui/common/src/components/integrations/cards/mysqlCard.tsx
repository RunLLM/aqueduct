import { Typography } from '@mui/material';
import Box from '@mui/material/Box';
import React from 'react';
import {Integration, MySqlConfig} from "../../../utils/integrations";

type Props = {
    integration: Integration;
};

export const MySqlCard: React.FC<Props> = ({ integration }) => {
    const config = integration.config as MySqlConfig;
    return (
        <Box sx={{ display: 'flex', flexDirection: 'column' }}>
            <Typography variant="body1">
                <strong>Host: </strong>
                {config.host}
            </Typography>
            <Typography variant="body1">
                <strong>Port: </strong>
                {config.port}
            </Typography>
            <Typography variant="body1">
                <strong>User: </strong>
                {config.username}
            </Typography>
            <Typography variant="body1">
                <strong>Database: </strong>
                {config.database}
            </Typography>
        </Box>
    );
};
