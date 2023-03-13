import { faCode } from '@fortawesome/free-solid-svg-icons';
import React, { memo } from 'react';

import { ReactFlowNodeData } from '../../../utils/reactflow';
import Node from './Node';
import { operatorNodeStatusLabels } from './nodeTypes';

type Props = {
  data: ReactFlowNodeData;
  isConnectable: boolean;
};

export const functionOperatorNodeIcon = faCode;

const FunctionOperatorNode: React.FC<Props> = ({ data, isConnectable }) => {
  return (
    <Node
      icon={functionOperatorNodeIcon}
      data={data}
      isConnectable={isConnectable}
      defaultLabel="Function"
      statusLabels={operatorNodeStatusLabels}
    />
  );
};

export default memo(FunctionOperatorNode);
