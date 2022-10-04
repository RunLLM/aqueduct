import { faHashtag } from '@fortawesome/free-solid-svg-icons';
import React, { memo } from 'react';

import { ReactFlowNodeData } from '../../../utils/reactflow';
import Node from './Node';

type Props = {
  data: ReactFlowNodeData;
  isConnectable: boolean;
};

const NumericArtifactNode: React.FC<Props> = ({ data, isConnectable }) => {
  return (
    <Node
      icon={faHashtag}
      data={data}
      isConnectable={isConnectable}
      defaultLabel="Numeric"
    />
  );
};

export default memo(NumericArtifactNode);
