import { faCircleCheck } from '@fortawesome/free-solid-svg-icons';
import React, { memo } from 'react';

import Node from './Node';
import {ReactFlowNodeData} from "../../../utils/reactflow";

type Props = {
    data: ReactFlowNodeData;
    isConnectable: boolean;
};

export const BoolArtifactNode: React.FC<Props> = ({ data, isConnectable }) => {
    return <Node icon={faCircleCheck} data={data} isConnectable={isConnectable} defaultLabel="Float" />;
};

export default memo(BoolArtifactNode);
