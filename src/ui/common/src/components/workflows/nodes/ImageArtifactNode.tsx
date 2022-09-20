import { faImage } from '@fortawesome/free-solid-svg-icons';
import React, { memo } from 'react';

import { ReactFlowNodeData } from '../../../utils/reactflow';
import Node from './Node';

type Props = {
  data: ReactFlowNodeData;
  isConnectable: boolean;
};

export const imageArtifactNodeIcon = faImage;

const ImageArtifactNode: React.FC<Props> = ({ data, isConnectable }) => {
  return (
    <Node
      icon={imageArtifactNodeIcon}
      data={data}
      isConnectable={isConnectable}
      defaultLabel="Image"
    />
  );
};

export default memo(ImageArtifactNode);
