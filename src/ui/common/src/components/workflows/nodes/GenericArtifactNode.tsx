import { faFile } from '@fortawesome/free-solid-svg-icons';
import React, { memo } from 'react';

import { ReactFlowNodeData } from '../../../utils/reactflow';
import Node from './Node';

type Props = {
  data: ReactFlowNodeData;
  isConnectable: boolean;
};

export const genericArtifactNodeIcon = faFile;

const GenericArtifactNode: React.FC<Props> = ({ data, isConnectable }) => {
  return (
    <Node
      icon={genericArtifactNodeIcon}
      data={data}
      isConnectable={isConnectable}
      defaultLabel="Artifact"
    />
  );
};

export default memo(GenericArtifactNode);
