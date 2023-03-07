import { faFileCode } from '@fortawesome/free-solid-svg-icons';
import React, { memo } from 'react';

import { ReactFlowNodeData } from '../../../utils/reactflow';
import Node from './Node';
import { artifactNodeStatusLabels } from './nodeTypes';

type Props = {
  data: ReactFlowNodeData;
  isConnectable: boolean;
};

export const dictArtifactNodeIcon = faFileCode;

const DictArtifactNode: React.FC<Props> = ({ data, isConnectable }) => {
  return (
    <Node
      icon={dictArtifactNodeIcon}
      data={data}
      isConnectable={isConnectable}
      defaultLabel="Dictionary"
      statusLabels={artifactNodeStatusLabels}
    />
  );
};

export default memo(DictArtifactNode);
