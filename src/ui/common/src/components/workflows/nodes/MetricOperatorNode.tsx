import { faTemperatureHalf } from '@fortawesome/free-solid-svg-icons';
import React, { memo } from 'react';

import { ReactFlowNodeData } from '../../../utils/reactflow';
import Node from './Node';
import { metricNodeStatusLabels } from './nodeTypes';

type Props = {
  data: ReactFlowNodeData;
  isConnectable: boolean;
};

export const metricOperatorNodeIcon = faTemperatureHalf;

export const parseMetricResult = (
  metricValue: string,
  sigfigs: number
): string => {
  // Check if the number passed in is a whole number, return that if so.
  const parsedFloat = parseFloat(metricValue);
  if (parsedFloat % 1 === 0) {
    return metricValue;
  }
  // Only show three decimal points.
  return parsedFloat.toFixed(sigfigs);
};

const MetricOperatorNode: React.FC<Props> = ({ data, isConnectable }) => {
  return (
    <Node
      icon={metricOperatorNodeIcon}
      data={data}
      isConnectable={isConnectable}
      defaultLabel="Metric"
      preview={
        data.result !== undefined ? parseMetricResult(data.result, 3) : '-'
      }
      statusLabels={metricNodeStatusLabels}
    />
  );
};

export default memo(MetricOperatorNode);
