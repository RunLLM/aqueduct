import { faMagnifyingGlass } from '@fortawesome/free-solid-svg-icons';
import React, { memo } from 'react';

import Node from './Node';
import {ReactFlowNodeData} from "../../../utils/reactflow";

type Props = {
    data: ReactFlowNodeData;
    isConnectable: boolean;
};

export const CheckOperatorNode: React.FC<Props> = ({ data, isConnectable }) => {
    return <Node icon={faMagnifyingGlass} data={data} isConnectable={isConnectable} defaultLabel="Check" />;
};

export default memo(CheckOperatorNode);
