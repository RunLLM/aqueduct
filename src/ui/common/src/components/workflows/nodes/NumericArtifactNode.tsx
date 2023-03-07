import { faHashtag } from '@fortawesome/free-solid-svg-icons';
import React, { memo } from 'react';

import { ReactFlowNodeData } from '../../../utils/reactflow';
import Node from './Node';
import { artifactNodeStatusLabels } from './nodeTypes';

type Props = {
  data: ReactFlowNodeData;
  isConnectable: boolean;
};

export const numericArtifactNodeIcon = faHashtag;

const NumericArtifactNode: React.FC<Props> = ({ data, isConnectable }) => {
  return (
    <Node
      icon={numericArtifactNodeIcon}
      data={data}
      isConnectable={isConnectable}
      defaultLabel="Numeric"
      statusLabels={artifactNodeStatusLabels}
    />
  );
};

export default memo(NumericArtifactNode);
