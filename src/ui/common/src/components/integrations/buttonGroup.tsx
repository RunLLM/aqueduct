import { faFileCsv, faFlask } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import Tooltip from '@mui/material/Tooltip';
import React from 'react';

import { Integration } from '../../utils/integrations';
import ButtonGroup from '../primitives/ButtonGroup';
import { IconButton } from '../primitives/IconButton.styles';

type Props = {
  integration: Integration;
  onUploadCsv?: () => void;
  onTestConnection?: () => void;
};

const IntegrationButtonGroup: React.FC<Props> = ({
  integration,
  onTestConnection,
  onUploadCsv,
}) => {
  return (
    <ButtonGroup>
      <Tooltip title="test-connect this integration">
        <IconButton sx={{ marginRight: '2px' }} onClick={onTestConnection}>
          <FontAwesomeIcon icon={faFlask} color="black" />
        </IconButton>
      </Tooltip>
      {integration.name === 'aqueduct_demo' && (
        <Tooltip title="upload a csv file">
          <IconButton>
            <FontAwesomeIcon
              icon={faFileCsv}
              onClick={onUploadCsv}
              color="black"
            />
          </IconButton>
        </Tooltip>
      )}
    </ButtonGroup>
  );
};

export default IntegrationButtonGroup;
