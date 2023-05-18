import { AlertTitle, Box, CircularProgress } from '@mui/material';
import Alert from '@mui/material/Alert';
import Typography from '@mui/material/Typography';
import Image from 'mui-image';
import React from 'react';

import { ArtifactResultResponse } from '../../../handlers/responses/node';
import {
  ArtifactResultContent,
  SerializationType,
} from '../../../utils/artifacts';
import { Data, inferSchema, TableRow } from '../../../utils/data';
import PaginatedTable from '../../tables/PaginatedTable';

type Props = {
  artifactResult?: ArtifactResultResponse;
  content?: ArtifactResultContent;
  contentLoading: boolean;
  contentError: string;
};

const ArtifactContent: React.FC<Props> = ({
  artifactResult,
  content,
  contentLoading,
  contentError,
}) => {
  // intentional '!=' check for null or undefined.
  if (artifactResult?.content_serialized != null) {
    return (
      <Typography variant="body1" component="div" marginBottom="8px">
        <code>{artifactResult.content_serialized}</code>
      </Typography>
    );
  }

  if (contentLoading) {
    return <CircularProgress />;
  }

  if (contentError) {
    return (
      <Alert severity="error">
        <AlertTitle>Failed to load artifact contents.</AlertTitle>
        {contentError}
      </Alert>
    );
  }

  if (!content || !artifactResult) {
    return (
      <Typography variant="h5" component="div" marginBottom="8px">
        No result to show for this artifact.
      </Typography>
    );
  }

  let contentComponent = null;
  switch (artifactResult.serialization_type) {
    case SerializationType.Table:
    case SerializationType.BsonTable:
      try {
        const rawData = JSON.parse(content.data);
        if (artifactResult.serialization_type === SerializationType.BsonTable) {
          const rows = rawData as TableRow[];
          // bson table does not include schema when serialized.
          const schema = inferSchema(rows);
          contentComponent = (
            <PaginatedTable data={{ schema: schema, data: rows }} />
          );
          break;
        }

        contentComponent = <PaginatedTable data={rawData as Data} />;
        break;
      } catch (err) {
        return (
          <Alert severity="error" title="Cannot parse table data.">
            {`${err.toString}\n${content.data}`}
          </Alert>
        );
      }
    case SerializationType.Image:
      try {
        const srcFromBase64 = 'data:image/png;base64,' + content.data;
        contentComponent = (
          <Image
            src={srcFromBase64}
            duration={0}
            fit="contain"
            width="fit-content"
          />
        );
        break;
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
        const prettyJson = JSON.stringify(JSON.parse(content.data), null, 2);
        contentComponent = (
          <Typography sx={{ fontFamily: 'Monospace', whiteSpace: 'pre-wrap' }}>
            {prettyJson}
          </Typography>
        );
        break;
      } catch (err) {
        return (
          <Alert severity="error" title="Cannot parse json data.">
            {err.toString()}
          </Alert>
        );
      }
    case SerializationType.String:
      contentComponent = (
        <Typography sx={{ fontFamily: 'Monospace', whiteSpace: 'pre-wrap' }}>
          {content.data}
        </Typography>
      );
      break;
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
            {artifactResult.serialization_type}.
          </Typography>
        </Alert>
      );
  }

  if (!content.is_downsampled) {
    return contentComponent;
  }

  return (
    <Box>
      <Alert severity="info" sx={{ marginBottom: 1 }}>
        Original content too large. Loading subset of data for preview.
      </Alert>
      {contentComponent}
    </Box>
  );
};

export default ArtifactContent;
