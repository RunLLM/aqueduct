import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React from 'react';

import {
  getLoadParametersType,
  LoadParameters,
  LoadParametersType,
  RelationalDBLoadParams,
  S3LoadParams,
} from '../../../utils/operators';
import RelationalDBLoadParamDisplay from './RelationalDBLoadParamDisplay';
import S3LoadParamDisplay from './S3LoadParamDisplay';

type SaveDetailsProps = {
  parameters: LoadParameters;
};

export const SaveDetails: React.FC<SaveDetailsProps> = ({ parameters }) => {
  let paramsDisplay = null;
  if (parameters) {
    switch (getLoadParametersType(parameters)) {
      case LoadParametersType.RelationalDBLoadParamsType:
        paramsDisplay = (
          <RelationalDBLoadParamDisplay
            parameters={parameters as RelationalDBLoadParams}
          />
        );
        break;
      case LoadParametersType.S3LoadParamsType:
        paramsDisplay = (
          <S3LoadParamDisplay parameters={parameters as S3LoadParams} />
        );
        break;
      default:
        return null;
    }
    return (
      <Box mb={2}>
        <Typography variant="h6" mb="8px" fontWeight="normal">
          Parameters
        </Typography>
        {paramsDisplay}
      </Box>
    );
  }
};

export default SaveDetails;
