import { faChartLine } from '@fortawesome/free-solid-svg-icons';
import React, { memo } from 'react';

import { ReactFlowNodeData } from '../../../utils/reactflow';
import Node from './Node';

type Props = {
  data: ReactFlowNodeData;
  isConnectable: boolean;
};

const FloatArtifactNode: React.FC<Props> = ({ data, isConnectable }) => {
  return (
    <Node
      icon={faChartLine}
      data={data}
      isConnectable={isConnectable}
      defaultLabel="Float"
    />
  );
};

export default memo(FloatArtifactNode);
