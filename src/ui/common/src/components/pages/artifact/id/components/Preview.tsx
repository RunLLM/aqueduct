import { Alert, Box, Divider, Typography } from '@mui/material';
import React from 'react';

import ArtifactContent from '../../../../../components/workflows/artifact/content';
import { ArtifactResultResponse } from '../../../../../handlers/responses/artifact';
import { ContentWithLoadingStatus } from '../../../../../reducers/artifactResultContents';

type PreviewProps = {
  upstreamPending: boolean;
  previewAvailable: boolean;
  artifact: ArtifactResultResponse;
  contentWithLoadingStatus: ContentWithLoadingStatus;
};

export const Preview: React.FC<PreviewProps> = ({
  upstreamPending,
  previewAvailable,
  artifact,
  contentWithLoadingStatus,
}) => {
  let preview = (
    <>
      <Divider sx={{ marginY: '32px' }} />

      <Box marginBottom="32px">
        <Alert severity="warning">
          An upstream operator failed, causing this artifact to not be created.
        </Alert>
      </Box>
    </>
  );

  if (upstreamPending) {
    preview = (
      <>
        <Divider sx={{ marginY: '32px' }} />

        <Box marginBottom="32px">
          <Alert severity="warning">
            An upstream operator is in progress so this artifact is not yet
            created.
          </Alert>
        </Box>
      </>
    );
  } else if (previewAvailable) {
    preview = (
      <>
        <Divider sx={{ marginY: '32px' }} />
        <Box width="100%" marginTop="12px">
          <Typography
            variant="h6"
            component="div"
            marginBottom="8px"
            fontWeight="normal"
          >
            Preview
          </Typography>
          <ArtifactContent
            artifact={artifact}
            contentWithLoadingStatus={contentWithLoadingStatus}
          />
        </Box>

        <Divider sx={{ marginY: '32px' }} />
      </>
    );
  }
  return preview;
};

export default Preview;
