import React from 'react';

import { ArtifactResponse } from '../../../handlers/responses/node';
import { NodeArtifactResultContentGetResponse } from '../../../handlers/v2/NodeArtifactResultContentGet';
import { ArtifactType } from '../../../utils/artifacts';
import { exportCsv } from '../../../utils/preview';
import { Button } from '../../primitives/Button.styles';
import { LoadingButton } from '../../primitives/LoadingButton.styles';

type Props = {
  artifact: ArtifactResponse;
  content?: NodeArtifactResultContentGetResponse;
  isLoading: boolean;
};

// CsvExporter returns a CSV download button if the artifact is exportable.
// Otherwise it returns `null`.
const CsvExporter: React.FC<Props> = ({ artifact, content, isLoading }) => {
  if (artifact.type !== ArtifactType.Table) {
    return null;
  }

  if (isLoading) {
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

  if (!content) {
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
          JSON.parse(content.content),
          artifact.name ? artifact.name.replaceAll(' ', '_') : 'data'
        );
      }}
    >
      Export
    </Button>
  );
};

export default CsvExporter;
