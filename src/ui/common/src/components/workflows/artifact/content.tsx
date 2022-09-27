import { CircularProgress } from '@mui/material';
import Alert from '@mui/material/Alert';
import Typography from '@mui/material/Typography';
import React from 'react';

import { ArtifactResultResponse } from '../../../handlers/responses/artifact';
import { ContentWithLoadingStatus } from '../../../reducers/artifactResultContents';
import { ArtifactType } from '../../../utils/artifacts';
import { isFailed, isInitial, isLoading } from '../../../utils/shared';
import { Button } from '../../primitives/Button.styles';
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
      <Alert title="Failed to load artifact contents.">
        {contentWithLoadingStatus.status.err}
      </Alert>
    );
  }

  if (!contentWithLoadingStatus.data) {
    return (
      <Typography variant="h5" component="div" marginBottom="8px">
        No result to show for this artifact.
      </Typography>
    );
  }

  if (
    artifact.type === ArtifactType.Bytes ||
    artifact.type === ArtifactType.Picklable
  ) {
    return (
      <Button
        variant="contained"
        sx={{ maxHeight: '32px' }}
        onClick={() => {
          const content = contentWithLoadingStatus.data;
          const blob = new Blob([content], { type: 'text' });
          const url = window.URL.createObjectURL(blob);
          const a = document.createElement('a');
          a.href = url;
          a.download = artifact.name;
          a.click();

          return true;
        }}
      >
        Download
      </Button>
    );
  }

  if (artifact.type === ArtifactType.Table) {
    try {
      const data = JSON.parse(contentWithLoadingStatus.data);
      return <PaginatedTable data={data} />;
    } catch (err) {
      return (
        <Alert title="Cannot parse table data.">
          {err}
          {contentWithLoadingStatus.data}
        </Alert>
      );
    }
  }
  // TODO: handle images here
  return (
    <Typography variant="body1" component="div" marginBottom="8px">
      <code>{contentWithLoadingStatus.data}</code>
    </Typography>
  );
};

export default ArtifactContent;
