import { Alert, Typography } from '@mui/material';
import Box from '@mui/material/Box';
import CircularProgress from '@mui/material/CircularProgress';
import React from 'react';

import { ObjectState } from '../../reducers/integration';
import { isFailed, isLoading, isSucceeded } from '../../utils/shared';
import VirtualizedTable from '../tables/virtualizedTable';

type Props = {
  objectName: string;
  object: ObjectState;
};

const IntegrationObjectPreview: React.FC<Props> = ({ objectName, object }) => {
  let content: React.ReactElement;
  if (isLoading(object.status)) {
    content = (
      <Box sx={{ display: 'flex', flexDirection: 'row', mt: 3 }}>
        <CircularProgress size={30} />
        <Typography sx={{ ml: 2 }}>
          Loading object <b>{objectName}</b>...
        </Typography>
      </Box>
    );
  }
  if (isFailed(object.status)) {
    content = (
      <Alert style={{ marginTop: '10px' }} severity="error">
        Object <b>{objectName}</b> failed to load. Try refreshing the page.{' '}
        <br />
        Error: {object.status.err}
      </Alert>
    );
  }
  if (isSucceeded(object.status) && !!object.data) {
    const columnsContent = object.data.schema.fields.map((column) => {
      return {
        dataKey: column.name,
        label: column.name,
        type: column.type,
      };
    });
    content = (
      <Box
        sx={{
          height: '50vh',
          width: '100%',
          overflow: 'auto',
          overflowY: 'hidden',
        }}
      >
        <VirtualizedTable
          rowCount={object.data.data.length}
          rowGetter={({ index }) => object.data.data[index]}
          columns={columnsContent}
        />
      </Box>
    );
  }
  return <Box sx={{ mt: 3 }}>{content}</Box>;
};
/*return (
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
};*/

export default IntegrationObjectPreview;
