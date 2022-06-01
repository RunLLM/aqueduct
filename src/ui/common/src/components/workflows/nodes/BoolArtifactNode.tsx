import { faCircleCheck } from '@fortawesome/free-solid-svg-icons';
import React, { memo } from 'react';

import { ReactFlowNodeData } from '../../../utils/reactflow';
import Node from './Node';

type Props = {
  data: ReactFlowNodeData;
  isConnectable: boolean;
};

export const BoolArtifactNode: React.FC<Props> = ({ data, isConnectable }) => {
  return (
    <Node
      icon={faCircleCheck}
      data={data}
      isConnectable={isConnectable}
      defaultLabel="Float"
    />
  );
};

export default memo(BoolArtifactNode);
