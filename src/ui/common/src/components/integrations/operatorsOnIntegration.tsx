import Alert from '@mui/material/Alert';
import AlertTitle from '@mui/material/AlertTitle';
import Box from '@mui/material/Box';
import React from 'react';
import { useSelector } from 'react-redux';

import { OperatorsForIntegrationItem } from '../../reducers/integration';
import { RootState } from '../../stores/store';
import { Integration } from '../../utils/integrations';
import { isFailed, isInitial } from '../../utils/shared';
import { ListWorkflowSummary } from '../../utils/workflows';
import WorkflowSummaryCard from '../workflows/WorkflowSummaryCard';

type OperatorsOnIntegrationProps = {
  integration: Integration;
};

const OperatorsOnIntegration: React.FC<OperatorsOnIntegrationProps> = ({
  integration,
}) => {
  const listWorkflowState = useSelector(
    (state: RootState) => state.listWorkflowReducer
  );
  const operatorsState = useSelector((state: RootState) => {
    return state.integrationReducer.operators;
  });

  if (
    isInitial(operatorsState.status) ||
    isInitial(listWorkflowState.loadingStatus)
  ) {
    return null;
  }

  if (
    isFailed(operatorsState.status) ||
    isFailed(listWorkflowState.loadingStatus)
  ) {
    return (
      <Alert severity="error">
        <AlertTitle>
          {
            "We couldn't retrieve workflows associated with this integration for now."
          }
        </AlertTitle>
      </Alert>
    );
  }

  const workflows = listWorkflowState.workflows;
  const operators = operatorsState.operators;

  const workflowMap: { [id: string]: ListWorkflowSummary } = {};
  workflows.map((wf) => {
    workflowMap[wf.id] = wf;
  });

  const operatorsByWorkflow: {
    [id: string]: {
      workflow?: ListWorkflowSummary;
      operators: OperatorsForIntegrationItem[];
    };
  } = {};
  operators.map((op) => {
    const wf = workflowMap[op.workflow_id];
    if (!operatorsByWorkflow[op.workflow_id]) {
      operatorsByWorkflow[op.workflow_id] = { workflow: wf, operators: [] };
    }
    operatorsByWorkflow[op.workflow_id].operators.push(op);
  });

  if (Object.keys(operatorsByWorkflow).length > 0) {
    return (
      <Box maxWidth="900px" sx={{ display: 'flex', flexWrap: 'wrap' }}>
        {Object.entries(operatorsByWorkflow).map(([wfId, item]) => {
          return (
            <WorkflowSummaryCard
              integration={integration}
              key={wfId}
              workflow={item.workflow}
              operators={operatorsByWorkflow[wfId].operators}
            />
          );
        })}
      </Box>
    );
  }

  return <Box>This resource is not used by any workflows.</Box>;
};

export default OperatorsOnIntegration;
