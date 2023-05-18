import { Alert, Box, Divider, Typography } from '@mui/material';
import React from 'react';

import ArtifactContent from '../../../../../components/workflows/artifact/content';
import { ArtifactResultResponse } from '../../../../../handlers/responses/node';
import { ArtifactResultContent } from '../../../../../utils/artifacts';

type PreviewProps = {
  upstreamPending: boolean;
  previewAvailable: boolean;
  artifactResult?: ArtifactResultResponse;
  content: ArtifactResultContent;
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
            artifactResult={artifactResult}
            content={content}
            contentLoading={contentLoading}
            contentError={contentError}
          />
        </Box>

        <Divider sx={{ marginY: '32px' }} />
      </>
    );
  }
  return preview;
};

export default Preview;
