import { faChartLine } from '@fortawesome/free-solid-svg-icons';
import {ReactFlowNodeData} from "../../../utils/reactflow";
import React, { memo } from 'react';

import Node from './Node';

type Props = {
    data: ReactFlowNodeData;
    isConnectable: boolean;
};

export const FloatArtifactNode: React.FC<Props> = ({ data, isConnectable }) => {
    return <Node icon={faChartLine} data={data} isConnectable={isConnectable} defaultLabel="Float" />;
};

export default memo(FloatArtifactNode);
