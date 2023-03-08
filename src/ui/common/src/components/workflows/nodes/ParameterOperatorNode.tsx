import { faSliders } from '@fortawesome/free-solid-svg-icons';
import React, { memo } from 'react';

import { ReactFlowNodeData } from '../../../utils/reactflow';
import Node from './Node';
import { artifactNodeStatusLabels } from './nodeTypes';

type Props = {
  data: ReactFlowNodeData;
  isConnectable: boolean;
};

export const paramOperatorNodeIcon = faSliders;

const ParameterOperatorNode: React.FC<Props> = ({ data, isConnectable }) => {
  return (
    <Node
      icon={paramOperatorNodeIcon}
      data={data}
      isConnectable={isConnectable}
      defaultLabel="Parameter"
      statusLabels={artifactNodeStatusLabels}
    />
  );
};

export default memo(ParameterOperatorNode);
