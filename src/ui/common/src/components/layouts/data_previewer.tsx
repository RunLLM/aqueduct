import Alert from '@mui/material/Alert';
import AlertTitle from '@mui/material/AlertTitle';
import Box from '@mui/material/Box';
import CircularProgress from '@mui/material/CircularProgress';
import Image from 'mui-image';
import React from 'react';

import { ArtifactResult } from '../../reducers/workflow';
import { SerializationType } from '../../utils/artifacts';
import { ExecutionStatus, LoadingStatusEnum } from '../../utils/shared';
import { Error } from '../../utils/shared';
import DataTable from '../tables/DataTable';
import LogBlock, { LogLevel } from '../text/LogBlock';
import TextBlock from '../text/TextBlock';

type Props = {
  previewData: ArtifactResult;
  error?: Error;
};

const DataPreviewer: React.FC<Props> = ({ previewData, error }) => {
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
          <TextBlock text={loadingStatus.err} />
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
              height: '100%',
              width: '100%',
              overflow: 'auto',
              overflowY: 'hidden',
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
        data = <Image src={srcFromBase64} duration={0} fit="contain" />;
        break;
      case SerializationType.Json:
        // Convert to pretty-printed version.
        const prettyJson = JSON.stringify(
          JSON.parse(previewData.result.data),
          null,
          2
        );
        data = <TextBlock text={prettyJson} />;
        break;
      case SerializationType.String:
        data = <TextBlock text={previewData.result.data} />;
        break;
      default:
        errorComponent = (
          <Box>
            <Alert severity="warning">
              <TextBlock text="Artifact contains binary data that cannot be previewed." />
            </Alert>
          </Box>
        );
    }
  }

  return (
    <>
      {errorComponent}
      {data}
    </>
  );
};

export default DataPreviewer;
