import Box from '@mui/material/Box';
import React from 'react';
import { useSelector } from 'react-redux';

import { RootState } from '../../../stores/store';
import { getUpstreamOperator } from '../../../utils/artifacts';
import DataPreviewer from '../../layouts/data_previewer';

type Props = {
  artifactId: string;
};

const DataPreviewSideSheet: React.FC<Props> = ({ artifactId }) => {
  const workflowState = useSelector(
    (state: RootState) => state.workflowReducer
  );
  const operatorId = getUpstreamOperator(workflowState, artifactId);
  const artifactResult = useSelector(
    (state: RootState) => state.workflowReducer.artifactResults[artifactId]
  );

  // Check to see if there was an error, and if there was, pull it out of the
  // operator state. The reason we need this code is because artifacts don't
  // know what operators they are associated with.
  const operatorResult = useSelector(
    (state: RootState) => state.workflowReducer.operatorResults[operatorId]
  );
  const error = operatorResult?.result?.error;

  return (
    <Box p={1} sx={{ height: '75%', overflow: 'auto' }}>
      <DataPreviewer previewData={artifactResult} error={error} />
    </Box>
  );
};

export default DataPreviewSideSheet;
