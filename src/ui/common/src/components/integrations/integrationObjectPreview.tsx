import { Alert, Typography } from '@mui/material';
import Box from '@mui/material/Box';
import CircularProgress from '@mui/material/CircularProgress';
import React from 'react';

import DataTable from '../../components/tables/dataTable';
import { ObjectState } from '../../reducers/integration';
import {
  isFailed,
  isInitial,
  isLoading,
  isSucceeded,
} from '../../utils/shared';

type Props = {
  objectName: string;
  object: ObjectState;
};

const IntegrationObjectPreview: React.FC<Props> = ({ objectName, object }) => {
  if (isInitial(object.status)) {
    return null;
  }

  return (
    <Box sx={{ mt: 3 }}>
      {isLoading(object.status) && (
        <Box sx={{ display: 'flex', flexDirection: 'row', mt: 3 }}>
          <CircularProgress size={30} />
          <Typography sx={{ ml: 2 }}>
            Loading object <b>{objectName}</b>...
          </Typography>
        </Box>
      )}
      {isFailed(object.status) && (
        <Alert style={{ marginTop: '10px' }} severity="error">
          Object <b>{objectName}</b> failed to load. Try refreshing the page.{' '}
          <br />
          Error: {object.status.err}
        </Alert>
      )}
      {isSucceeded(object.status) && !!object.data && (
        <Box sx={{ height: '50vh', width: '100%' }}>
          <DataTable data={object.data} />
        </Box>
      )}
    </Box>
  );
};

export default IntegrationObjectPreview;
