import { faMagnifyingGlass } from '@fortawesome/free-solid-svg-icons';
import React, { memo } from 'react';

import { ReactFlowNodeData } from '../../../utils/reactflow';
import Node from './Node';

type Props = {
  data: ReactFlowNodeData;
  isConnectable: boolean;
};

export const checkOperatorNodeIcon = faMagnifyingGlass;

const CheckOperatorNode: React.FC<Props> = ({ data, isConnectable }) => {
  return (
    <Node
      icon={checkOperatorNodeIcon}
      data={data}
      isConnectable={isConnectable}
      defaultLabel="Check"
    />
  );
};

export default memo(CheckOperatorNode);
