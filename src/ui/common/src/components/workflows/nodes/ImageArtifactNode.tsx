import { faImage } from '@fortawesome/free-solid-svg-icons';
import React, { memo } from 'react';

import { ReactFlowNodeData } from '../../../utils/reactflow';
import Node from './Node';
import { artifactNodeStatusLabels } from './nodeTypes';

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
      statusLabels={artifactNodeStatusLabels}
    />
  );
};

export default memo(ImageArtifactNode);
