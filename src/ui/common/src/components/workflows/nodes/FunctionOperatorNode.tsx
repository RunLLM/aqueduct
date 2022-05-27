import { faCode } from '@fortawesome/free-solid-svg-icons';
import {ReactFlowNodeData} from "../../../utils/reactflow";
import React, { memo } from 'react';

import Node from './Node';

type Props = {
    data: ReactFlowNodeData;
    isConnectable: boolean;
};

export const FunctionOperatorNode: React.FC<Props> = ({ data, isConnectable }) => {
    return <Node icon={faCode} data={data} isConnectable={isConnectable} defaultLabel="Function" />;
};

export default memo(FunctionOperatorNode);
