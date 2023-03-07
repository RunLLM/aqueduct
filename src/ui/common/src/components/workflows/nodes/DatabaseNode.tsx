import { faDatabase } from '@fortawesome/free-solid-svg-icons';
import React, { memo } from 'react';

import { ReactFlowNodeData } from '../../../utils/reactflow';
import Node from './Node';
import { operatorNodeStatusLabels } from './nodeTypes';

type Props = {
  data: ReactFlowNodeData;
  isConnectable: boolean;
};

export const databaseNodeIcon = faDatabase;

const DatabaseNode: React.FC<Props> = ({ data, isConnectable }) => {
  return (
    <Node
      icon={databaseNodeIcon}
      data={data}
      isConnectable={isConnectable}
      defaultLabel="Database"
      statusLabels={operatorNodeStatusLabels}
    />
  );
};

export default memo(DatabaseNode);
