import { Box, Tooltip } from '@mui/material';
import React from 'react';
import { IconDefinition, } from '@fortawesome/fontawesome-svg-core'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';


type TooltipIconProps = {
    icon: IconDefinition; // This icon is assumed to be a FontAwesome icon.
    tooltip: string;
    size?: string; // Assuming it's a square icon, width and height in px.
    color?: string; // MUI Theme compatible coloring.
};

export const TooltipIcon: React.FC<TooltipIconProps> = ({ icon, tooltip, size = '16px', color = 'black' }) => {
    return (
        <Tooltip placement="bottom" title={tooltip}>
            <Box sx={{
                width: size,
                height: size,
                fontSize: '16px',
                color={color}
            }}>
                <FontAwesomeIcon icon={icon} />
            </Box>
        </Tooltip>
    );
}