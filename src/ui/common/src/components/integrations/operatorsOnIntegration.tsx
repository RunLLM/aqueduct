import Alert from '@mui/material/Alert';
import AlertTitle from '@mui/material/AlertTitle';
import Box from '@mui/material/Box';
import CircularProgress from '@mui/material/CircularProgress';
import React, { useState } from 'react';
import { useSelector } from 'react-redux';

import { OperatorsForIntegrationItem } from '../../reducers/integration';
import { RootState } from '../../stores/store';
import { isFailed, isInitial, isLoading } from '../../utils/shared';
import { ListWorkflowSummary } from '../../utils/workflows';
import WorkflowAccordion from '../workflows/accordion';
import WorkflowSummaryCard from '../workflows/WorkflowSummaryCard';

const OperatorsOnIntegration: React.FC = () => {
  const listWorkflowState = useSelector(
    (state: RootState) => state.listWorkflowReducer
  );
  const operatorsState = useSelector((state: RootState) => {
    return state.integrationReducer.operators;
  });
  const [expandedWf, setExpandedWf] = useState<string>('');
  if (
    isInitial(operatorsState.status) ||
    isInitial(listWorkflowState.loadingStatus)
  ) {
    return null;
  }

  if (
    isLoading(operatorsState.status) ||
    isLoading(listWorkflowState.loadingStatus)
  ) {
    return <CircularProgress />;
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
      <>
        <Box maxWidth="900px">
          {Object.entries(operatorsByWorkflow).map(([wfId, item]) => {
            console.log(item);
            return (
              <WorkflowAccordion
                expanded={expandedWf === wfId}
                handleExpand={() => {
                  if (expandedWf === wfId) {
                    setExpandedWf('');
                    return;
                  }
                  setExpandedWf(wfId);
                }}
                key={wfId}
                workflow={item.workflow}
                operators={item.operators}
              />
            );
          })}
        </Box>

        <Box maxWidth="900px">
          {Object.entries(operatorsByWorkflow).map(([wfId, item]) => {
            console.log(item);
            return (
              <WorkflowSummaryCard
                expanded={expandedWf === wfId}
                handleExpand={() => {
                  if (expandedWf === wfId) {
                    setExpandedWf('');
                    return;
                  }
                  setExpandedWf(wfId);
                }}
                key={wfId}
                workflow={item.workflow}
                operators={item.operators}
              />
            );
          })}
        </Box>
      </>
    );
  } else {
    return <Box>This integration is not used by any workflows.</Box>;
  }
};

export default OperatorsOnIntegration;
