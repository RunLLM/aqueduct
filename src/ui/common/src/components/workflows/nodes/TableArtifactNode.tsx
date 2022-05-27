import { faTableColumns } from '@fortawesome/free-solid-svg-icons';
import React, { memo } from 'react';
import Node from './Node';
import {ReactFlowNodeData} from "../../../utils/reactflow";

type Props = {
    data: ReactFlowNodeData;
    isConnectable: boolean;
};

export const TableArtifactNode: React.FC<Props> = ({ data, isConnectable }) => {
    return <Node icon={faTableColumns} data={data} isConnectable={isConnectable} defaultLabel="Table" />;
};

export default memo(TableArtifactNode);
