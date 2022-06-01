import { faTemperatureHalf } from '@fortawesome/free-solid-svg-icons';
import React, { memo } from 'react';

import { ReactFlowNodeData } from '../../../utils/reactflow';
import Node from './Node';

type Props = {
  data: ReactFlowNodeData;
  isConnectable: boolean;
};

export const MetricOperatorNode: React.FC<Props> = ({
  data,
  isConnectable,
}) => {
  return (
    <Node
      icon={faTemperatureHalf}
      data={data}
      isConnectable={isConnectable}
      defaultLabel="Metric"
    />
  );
};

export default memo(MetricOperatorNode);
