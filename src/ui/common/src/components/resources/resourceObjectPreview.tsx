import { Alert, AlertTitle, Typography } from '@mui/material';
import Box from '@mui/material/Box';
import CircularProgress from '@mui/material/CircularProgress';
import React from 'react';

import { ObjectState } from '../../reducers/resource';
import {
  isFailed,
  isInitial,
  isLoading,
  isSucceeded,
} from '../../utils/shared';
import PaginatedTable from '../tables/PaginatedTable';

type Props = {
  objectName: string;
  object: ObjectState;
};

const ResourceObjectPreview: React.FC<Props> = ({ objectName, object }) => {
  if (!object || isInitial(object.status)) {
    return null;
  }

  let content: React.ReactElement;
  if (isLoading(object.status)) {
    content = (
      <Box sx={{ display: 'flex', flexDirection: 'row', mt: 3 }}>
        <CircularProgress size={30} />
        <Typography sx={{ ml: 2 }}>
          Loading <b>{objectName}</b>...
        </Typography>
      </Box>
    );
  }

  if (isFailed(object.status)) {
    content = (
      <Alert style={{ marginTop: '10px' }} severity="error">
        <AlertTitle>
          Object <b>{objectName}</b> failed to load.
        </AlertTitle>
        <pre>Error: {object.status.err}</pre>
      </Alert>
    );
  }

  if (isSucceeded(object.status) && !!object.data) {
    content = (
      <Box
        sx={{
          height: '50vh',
          width: '100%',
          overflow: 'auto',
          overflowY: 'hidden',
        }}
      >
        <PaginatedTable data={object.data} />
      </Box>
    );
  }
  return <Box sx={{ mt: 3 }}>{content}</Box>;
};

export default ResourceObjectPreview;
