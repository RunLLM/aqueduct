import { AlertTitle, CircularProgress } from '@mui/material';
import Alert from '@mui/material/Alert';
import Typography from '@mui/material/Typography';
import Image from 'mui-image';
import React from 'react';

import { ArtifactResultResponse } from '../../../handlers/responses/artifact';
import { ContentWithLoadingStatus } from '../../../reducers/artifactResultContents';
import { SerializationType } from '../../../utils/artifacts';
import { Data, inferSchema, TableRow } from '../../../utils/data';
import { isFailed, isInitial, isLoading } from '../../../utils/shared';
import PaginatedTable from '../../tables/PaginatedTable';

type Props = {
  artifact: ArtifactResultResponse;
  contentWithLoadingStatus?: ContentWithLoadingStatus;
};

const ArtifactContent: React.FC<Props> = ({
  artifact,
  contentWithLoadingStatus,
}) => {
  if (!artifact.result) {
    return (
      <Typography variant="h5" component="div" marginBottom="8px">
        No result to show for this artifact.
      </Typography>
    );
  }

  // intentional '!=' check for null or undefined.
  if (artifact.result.content_serialized != null) {
    return (
      <Typography variant="body1" component="div" marginBottom="8px">
        <code>{artifact.result.content_serialized}</code>
      </Typography>
    );
  }

  if (!contentWithLoadingStatus) {
    return <CircularProgress />;
  }
  if (
    isInitial(contentWithLoadingStatus.status) ||
    isLoading(contentWithLoadingStatus.status)
  ) {
    return <CircularProgress />;
  }

  if (isFailed(contentWithLoadingStatus.status)) {
    return (
      <Alert severity="error">
        <AlertTitle>Failed to load artifact contents.</AlertTitle>
        {contentWithLoadingStatus.status.err}
      </Alert>
    );
  }

  switch (artifact.result.serialization_type) {
    case SerializationType.Table:
    case SerializationType.BsonTable:
      try {
        const rawData = JSON.parse(contentWithLoadingStatus.data);
        if (
          artifact.result.serialization_type === SerializationType.BsonTable
        ) {
          const rows = rawData as TableRow[];
          const schema = inferSchema(rows);
          return <PaginatedTable data={{ schema: schema, data: rows }} />;
        }

        return <PaginatedTable data={rawData as Data} />;
      } catch (err) {
        return (
          <Alert severity="error" title="Cannot parse table data.">
            {`${err.toString}\n${contentWithLoadingStatus.data}`}
          </Alert>
        );
      }
    case SerializationType.Image:
      try {
        const srcFromBase64 =
          'data:image/png;base64,' + contentWithLoadingStatus.data;
        return (
          <Image
            src={srcFromBase64}
            duration={0}
            fit="contain"
            width="fit-content"
          />
        );
      } catch (err) {
        return (
          <Alert severity="error" title="Cannot parse image data.">
            {err}
          </Alert>
        );
      }
    case SerializationType.Json:
      try {
        // Convert to pretty-printed version.
        const prettyJson = JSON.stringify(
          JSON.parse(contentWithLoadingStatus.data),
          null,
          2
        );
        return (
          <Typography sx={{ fontFamily: 'Monospace', whiteSpace: 'pre-wrap' }}>
            {prettyJson}
          </Typography>
        );
      } catch (err) {
        return (
          <Alert severity="error" title="Cannot parse json data.">
            {err.toString()}
          </Alert>
        );
      }
    case SerializationType.String:
      return (
        <Typography sx={{ fontFamily: 'Monospace', whiteSpace: 'pre-wrap' }}>
          {contentWithLoadingStatus.data}
        </Typography>
      );
    case SerializationType.Bytes:
    case SerializationType.Pickle:
      return (
        <Alert severity="info">
          <Typography sx={{ whiteSpace: 'pre-wrap' }}>
            Artifact contains binary data that cannot be previewed.
          </Typography>
        </Alert>
      );
    default:
      return (
        <Alert severity="error">
          <Typography sx={{ whiteSpace: 'pre-wrap' }}>
            Cannot show preview due to unexpected serialization type:{' '}
            {artifact.result.serialization_type}.
          </Typography>
        </Alert>
      );
  }
};

export default ArtifactContent;
