import { faTemperatureHalf } from '@fortawesome/free-solid-svg-icons';
import React, { memo } from 'react';

import { ReactFlowNodeData } from '../../../utils/reactflow';
import Node from './Node';

type Props = {
  data: ReactFlowNodeData;
  isConnectable: boolean;
};

export const metricOperatorNodeIcon = faTemperatureHalf;

const MetricOperatorNode: React.FC<Props> = ({ data, isConnectable }) => {
  return (
    <Node
      icon={metricOperatorNodeIcon}
      data={data}
      isConnectable={isConnectable}
      defaultLabel="Metric"
    />
  );
};

export default memo(MetricOperatorNode);
