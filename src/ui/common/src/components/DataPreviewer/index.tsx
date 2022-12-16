import Alert from '@mui/material/Alert';
import AlertTitle from '@mui/material/AlertTitle';
import Box from '@mui/material/Box';
import CircularProgress from '@mui/material/CircularProgress';
import Typography from '@mui/material/Typography';
import Image from 'mui-image';
import React from 'react';

import { ArtifactResult } from '../../reducers/workflow';
import { SerializationType } from '../../utils/artifacts';
import { ExecutionStatus, LoadingStatusEnum } from '../../utils/shared';
import { Error } from '../../utils/shared';
import DataTable from '../tables/DataTable';
import LogBlock, { LogLevel } from '../text/LogBlock';

type Props = {
  previewData: ArtifactResult;
  error?: Error;
  // TODO: remove this if no longer needed.
  dataTableHeight?: string;
};

/**
 * Shows a preview for an artifact depending on it's serialization type.
 */
export const DataPreviewer: React.FC<Props> = ({
  previewData,
  error,
  dataTableHeight,
}) => {
  if (!previewData) {
    return null;
  }

  const errorMsg = !!error ? `${error.tip ?? ''}\n${error.context ?? ''}` : '';

  let errorComponent: React.ReactElement;
  if (error) {
    errorComponent = (
      <LogBlock
        logText={errorMsg.trim()}
        level={LogLevel.Error}
        title="An error occurred during execution"
      />
    );
  } else if (previewData.result?.status === ExecutionStatus.Failed) {
    // If the execution status is marked as failed but there is no error,
    // that means something upstream from where we are failed.
    errorComponent = (
      <Alert severity="warning">
        <AlertTitle>
          An upstream operator failed, ending the workflow execution.
        </AlertTitle>
      </Alert>
    );
  }

  const loadingStatus = previewData.loadingStatus;

  if (loadingStatus && loadingStatus.loading === LoadingStatusEnum.Loading) {
    return (
      <Box
        sx={{
          width: '100%',
          height: '100%',
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
        }}
      >
        <CircularProgress />
      </Box>
    );
  }

  if (loadingStatus && loadingStatus.loading === LoadingStatusEnum.Failed) {
    errorComponent = (
      <Box>
        <Alert severity="error">
          <Typography sx={{ whiteSpace: 'pre-wrap' }}>
            {loadingStatus.err}
          </Typography>
        </Alert>
      </Box>
    );
  }

  if (errorComponent) {
    return <>{errorComponent}</>;
  }

  let data: React.ReactElement;
  if (previewData.result?.status === ExecutionStatus.Succeeded) {
    switch (previewData.result.serialization_type) {
      case SerializationType.Table:
        const parsedData = JSON.parse(previewData.result.data);
        const columnsContent = parsedData.schema.fields.map((column) => {
          return {
            dataKey: column.name,
            label: column.name,
            type: column.type,
          };
        });
        data = (
          <Box
            sx={{
              position: 'absolute',
              height: dataTableHeight ? dataTableHeight : '100%',
            }}
          >
            <DataTable
              rowCount={parsedData.data.length}
              rowGetter={({ index }) => parsedData.data[index]}
              columns={columnsContent}
            />
          </Box>
        );
        break;
      case SerializationType.Image:
        const srcFromBase64 =
          'data:image/png;base64,' + previewData.result.data;
        data = (
          <Image
            src={srcFromBase64}
            duration={0}
            fit="contain"
            width="max-content"
          />
        );
        break;
      case SerializationType.Json:
        // Convert to pretty-printed version.
        const prettyJson = JSON.stringify(
          JSON.parse(previewData.result.data),
          null,
          2
        );
        data = (
          <Typography sx={{ fontFamily: 'Monospace', whiteSpace: 'pre-wrap' }}>
            {prettyJson}
          </Typography>
        );
        break;
      case SerializationType.String:
        data = (
          <Typography sx={{ fontFamily: 'Monospace', whiteSpace: 'pre-wrap' }}>
            {previewData.result.data}
          </Typography>
        );
        break;
      default:
        errorComponent = (
          <Box>
            <Alert severity="info">
              <Typography sx={{ whiteSpace: 'pre-wrap' }}>
                Artifact contains binary data that cannot be previewed.
              </Typography>
            </Alert>
          </Box>
        );
    }
  }

  return (
    <Box>
      {errorComponent}
      {data}
    </Box>
  );
};

export default DataPreviewer;
