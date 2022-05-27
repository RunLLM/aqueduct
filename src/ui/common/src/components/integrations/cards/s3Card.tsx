import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React from 'react';
import {S3Config} from "../../../utils/workflows";
import {Integration} from "../../../utils/integrations";

type Props = {
    integration: Integration;
};

export const S3Card: React.FC<Props> = ({ integration }) => {
    const config = integration.config as S3Config;
    return (
        <Box>
            <Typography variant="body1">
                <strong>Bucket: </strong>
                {config.bucket}
            </Typography>
        </Box>
    );
};
