import { Alert, Box, Typography } from '@mui/material';
import React from 'react';
import ExecutionStatus from '../../../../../utils/shared';

import ArtifactContent from '../../../../../components/workflows/artifact/content';
import { ArtifactResultResponse } from '../../../../../handlers/responses/node';
import { NodeArtifactResultContentGetResponse } from '../../../../../handlers/v2/NodeArtifactResultContentGet';

type PreviewProps = {
  upstreamPending: boolean;
  previewAvailable: boolean;
  artifactResult?: ArtifactResultResponse;
  content?: NodeArtifactResultContentGetResponse;
  contentLoading: boolean;
  contentError: string;
};

export const Preview: React.FC<PreviewProps> = ({
  upstreamPending,
  previewAvailable,
  artifactResult,
  content,
  contentLoading,
  contentError,
}) => {
  if (upstreamPending) {
    return (
      <Alert severity="warning">
        An upstream operator is in progress so this artifact is not yet created.
      </Alert>
    );
  }

  if (previewAvailable) {
    if (artifactResult?.exec_state?.status === ExecutionStatus.Deleted) {
      return (
        <Box marginBottom="32px">
          <Alert severity="info">
            This artifact has succeeded, but the snapshot has been deleted.
          </Alert>
        </Box>
      );
    }
    return (
      <Box width="100%">
        <Typography
          variant="h6"
          component="div"
          marginBottom="8px"
          fontWeight="normal"
        >
          Preview
        </Typography>
        <ArtifactContent
          artifactResult={artifactResult}
          content={content}
          contentLoading={contentLoading}
          contentError={contentError}
        />
      </Box>
    );
  }

  return (
    <Alert severity="warning">
      An upstream operator failed, causing this artifact to not be created.
    </Alert>
  );
};

export default Preview;
