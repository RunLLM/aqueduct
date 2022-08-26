import { faCircleCheck } from '@fortawesome/free-solid-svg-icons';
import React, { memo } from 'react';

import { ReactFlowNodeData } from '../../../utils/reactflow';
import Node from './Node';

type Props = {
  data: ReactFlowNodeData;
  isConnectable: boolean;
};

export const boolArtifactNodeIcon = faCircleCheck;

const BoolArtifactNode: React.FC<Props> = ({ data, isConnectable }) => {
  return (
    <Node
      icon={boolArtifactNodeIcon}
      data={data}
      isConnectable={isConnectable}
      defaultLabel="Bool"
    />
  );
};

export default memo(BoolArtifactNode);
