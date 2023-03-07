import { faTableColumns } from '@fortawesome/free-solid-svg-icons';
import React, { memo } from 'react';

import { ReactFlowNodeData } from '../../../utils/reactflow';
import Node from './Node';
import { artifactNodeStatusLabels } from './nodeTypes';

type Props = {
  data: ReactFlowNodeData;
  isConnectable: boolean;
};

export const tableArtifactNodeIcon = faTableColumns;

const TableArtifactNode: React.FC<Props> = ({ data, isConnectable }) => {
  return (
    <Node
      icon={tableArtifactNodeIcon}
      data={data}
      isConnectable={isConnectable}
      defaultLabel="Table"
      statusLabels={artifactNodeStatusLabels}
    />
  );
};

export default memo(TableArtifactNode);
