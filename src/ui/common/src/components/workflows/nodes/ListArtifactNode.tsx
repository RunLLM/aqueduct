import { faList } from '@fortawesome/free-solid-svg-icons';
import React, { memo } from 'react';

import { ReactFlowNodeData } from '../../../utils/reactflow';
import Node from './Node';

type Props = {
  data: ReactFlowNodeData;
  isConnectable: boolean;
};

export const listArtifactNodeIcon = faList;

const ListArtifactNode: React.FC<Props> = ({ data, isConnectable }) => {
  return (
    <Node
      icon={listArtifactNodeIcon}
      data={data}
      isConnectable={isConnectable}
      defaultLabel="List"
    />
  );
};

export default memo(ListArtifactNode);
