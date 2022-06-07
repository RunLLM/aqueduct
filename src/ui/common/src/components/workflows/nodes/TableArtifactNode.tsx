import { faTableColumns } from '@fortawesome/free-solid-svg-icons';
import React, { memo } from 'react';

import { ReactFlowNodeData } from '../../../utils/reactflow';
import Node from './Node';

type Props = {
  data: ReactFlowNodeData;
  isConnectable: boolean;
};

const TableArtifactNode: React.FC<Props> = ({ data, isConnectable }) => {
  return (
    <Node
      icon={faTableColumns}
      data={data}
      isConnectable={isConnectable}
      defaultLabel="Table"
    />
  );
};

export default memo(TableArtifactNode);
