import { faFileCode } from '@fortawesome/free-solid-svg-icons';
import React, { memo } from 'react';

import { ReactFlowNodeData } from '../../../utils/reactflow';
import Node from './Node';

type Props = {
  data: ReactFlowNodeData;
  isConnectable: boolean;
};

export const dictArtifactNodeIcon = faFileCode;

const DictArtifactNode: React.FC<Props> = ({ data, isConnectable }) => {
  return (
    <Node
      icon={dictArtifactNodeIcon}
      data={data}
      isConnectable={isConnectable}
      defaultLabel="Dictionary"
    />
  );
};

export default memo(DictArtifactNode);
