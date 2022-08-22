import {
    faCircleCheck,
    faCircleExclamation,
    faTriangleExclamation,
    faMinus,
} from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';

import React from 'react';
import { theme } from '../../styles/theme/theme';
import Box from '@mui/material/Box';

interface MetricTableItemProps {
    checkValue: string;
}

export const CheckTableItem: React.FC<MetricTableItemProps> = ({
    checkValue
}) => {
    let iconColor = theme.palette.black;
    let checkIcon = faMinus;

    switch (checkValue.toLowerCase()) {
        case 'true': {
            checkIcon = faCircleCheck;
            iconColor = theme.palette.green['400'];
            break;
        }
        case 'false': {
            checkIcon = faCircleExclamation;
            iconColor = theme.palette.red['500'];
            break;
        }
        case 'warning': {
            checkIcon = faTriangleExclamation;
            iconColor = theme.palette.orange['500'];
            break;
        }
        case 'none': {
            checkIcon = faMinus;
            iconColor = theme.palette.black;
            break;
        }
        default: {
            // None of the icon cases met, just fall through and render table value.
            return <>{checkValue}</>
        }
    }

    return (
        <Box sx={{ fontSize: '16px', color: iconColor }}>
            <FontAwesomeIcon icon={checkIcon} />
        </Box>
    );
}

export default CheckTableItem;
