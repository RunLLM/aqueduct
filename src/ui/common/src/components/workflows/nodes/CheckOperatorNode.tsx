import { faMagnifyingGlass } from '@fortawesome/free-solid-svg-icons';
import React, { memo } from 'react';
import { RootState } from '../../../stores/store';
import { useSelector } from 'react-redux';

import { ReactFlowNodeData } from '../../../utils/reactflow';
import Node from './Node';

type Props = {
  data: ReactFlowNodeData;
  isConnectable: boolean;
};

export const checkOperatorNodeIcon = faMagnifyingGlass;

const CheckOperatorNode: React.FC<Props> = ({ data, isConnectable }) => {
  // return (
  //   <Node
  //     icon={checkOperatorNodeIcon}
  //     data={data}
  //     isConnectable={isConnectable}
  //     defaultLabel="Check"
  //   />
  // );

  const operatorResults = useSelector((state: RootState) => state.workflowReducer.operatorResults);
  console.log('operatorResultData: ', operatorResults[data.nodeId]);
  const artifactResults = useSelector((state: RootState) => state.workflowReducer.artifactResults);
  console.log('artifactResultData', artifactResults[data.nodeId]);
  return (
    <Node
      data={data}
      isConnectable={isConnectable}
      defaultLabel="Check"
    />
  );
};

export default memo(CheckOperatorNode);
