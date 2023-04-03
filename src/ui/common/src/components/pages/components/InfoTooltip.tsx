import { faQuestionCircle } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { IconButton, Tooltip } from '@mui/material';
import React from 'react';

import { theme } from '../../../styles/theme/theme';

type TooltipPlacement =
  | 'right'
  | 'bottom-end'
  | 'bottom-start'
  | 'bottom'
  | 'left-end'
  | 'left-start'
  | 'left'
  | 'right-end'
  | 'right-start'
  | 'top-end'
  | 'top-start'
  | 'top';

type InfoTooltipProps = {
  /**
   * Text to show in tooltip.
   */
  tooltipText: string;
  /**
   * Placement of the tooltip relative to the (?) icon. Default is right.
   */
  placement?: TooltipPlacement;
};

/**
 * Infobutton where, when you hover, shows the `tooltipText` as a tooltip.
 **/
export const InfoTooltip: React.FC<InfoTooltipProps> = ({
  tooltipText,
  placement = 'right',
}) => {
  return (
    <Tooltip
      arrow
      placement={placement as TooltipPlacement}
      title={tooltipText}
    >
      <IconButton>
        <FontAwesomeIcon
          color={`${theme.palette.gray[700]}`}
          fontSize="16px"
          icon={faQuestionCircle}
        />
      </IconButton>
    </Tooltip>
  );
};
