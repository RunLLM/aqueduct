import { faChartLine } from '@fortawesome/free-solid-svg-icons';
import React, { memo } from 'react';

import { ReactFlowNodeData } from '../../../utils/reactflow';
import Node from './Node';

type Props = {
  data: ReactFlowNodeData;
  isConnectable: boolean;
};

const NoneArtifactNode: React.FC<Props> = ({ data, isConnectable }) => {
  console.log("node data: ", data)
  return (
    <Node
      icon={faChartLine}
      data={data}
      isConnectable={isConnectable}
      defaultLabel="None"
    />
  );
};

export default memo(NoneArtifactNode);