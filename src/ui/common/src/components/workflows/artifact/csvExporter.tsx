import React from 'react';

import { ArtifactResultResponse } from '../../../handlers/responses/artifactDeprecated';
import { ContentWithLoadingStatus } from '../../../reducers/artifactResultContents';
import { ArtifactType } from '../../../utils/artifacts';
import { exportCsv } from '../../../utils/preview';
import { isFailed, isInitial, isLoading } from '../../../utils/shared';
import { Button } from '../../primitives/Button.styles';
import { LoadingButton } from '../../primitives/LoadingButton.styles';

type Props = {
  artifact: ArtifactResultResponse;
  contentWithLoadingStatus?: ContentWithLoadingStatus;
};

// CsvExporter returns a CSV download button if the artifact is exportable.
// Otherwise it returns `null`.
const CsvExporter: React.FC<Props> = ({
  artifact,
  contentWithLoadingStatus,
}) => {
  if (artifact.type !== ArtifactType.Table) {
    return null;
  }

  if (!contentWithLoadingStatus) {
    return null;
  }

  if (
    isInitial(contentWithLoadingStatus.status) ||
    isLoading(contentWithLoadingStatus.status)
  ) {
    return (
      <LoadingButton
        variant="contained"
        sx={{ maxHeight: '32px' }}
        disabled
        loading
      >
        Export
      </LoadingButton>
    );
  }

  if (
    isFailed(contentWithLoadingStatus.status) ||
    !contentWithLoadingStatus.data
  ) {
    return (
      <Button variant="contained" sx={{ maxHeight: '32px' }} disabled>
        Export
      </Button>
    );
  }

  return (
    <Button
      variant="contained"
      sx={{ maxHeight: '32px' }}
      onClick={() => {
        exportCsv(
          JSON.parse(contentWithLoadingStatus.data),
          artifact.name ? artifact.name.replaceAll(' ', '_') : 'data'
        );
      }}
    >
      Export
    </Button>
  );
};

export default CsvExporter;
