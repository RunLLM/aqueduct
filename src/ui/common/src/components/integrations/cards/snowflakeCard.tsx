import { Typography } from '@mui/material';
import Box from '@mui/material/Box';
import React from 'react';
import {Integration, SnowflakeConfig} from "../../../utils/integrations";

type Props = {
    integration: Integration;
};

export const SnowflakeCard: React.FC<Props> = ({ integration }) => {
    const config = integration.config as SnowflakeConfig;
    return (
        <Box sx={{ display: 'flex', flexDirection: 'column' }}>
            <Typography variant="body1">
                <strong>Account Identifier: </strong>
                {config.account_identifier}
            </Typography>
            <Typography variant="body1">
                <strong>Warehouse: </strong>
                {config.warehouse}
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
