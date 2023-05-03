import Alert from '@mui/material/Alert';
import AlertTitle from '@mui/material/AlertTitle';
import Box from '@mui/material/Box';
import React from 'react';
import { useSelector } from 'react-redux';

import { OperatorResponse } from '../../handlers/responses/node';
import { RootState } from '../../stores/store';
import { Integration } from '../../utils/integrations';
import { isFailed, isInitial } from '../../utils/shared';
import { ListWorkflowSummary } from '../../utils/workflows';
import WorkflowSummaryCard from '../workflows/WorkflowSummaryCard';

type IntegrationWorkflowSummaryCardsProps = {
  integration: Integration;
  workflowIDToLatestOperators: { [workflowID: string]: OperatorResponse[] };
};

const IntegrationWorkflowSummaryCards: React.FC<
  IntegrationWorkflowSummaryCardsProps
> = ({ integration, workflowIDToLatestOperators }) => {
  const listWorkflowState = useSelector(
    (state: RootState) => state.listWorkflowReducer
  );

  if (isInitial(listWorkflowState.loadingStatus)) {
    return null;
  }

  if (isFailed(listWorkflowState.loadingStatus)) {
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
  const workflowMap: { [id: string]: ListWorkflowSummary } = {};
  workflows.map((wf) => {
    workflowMap[wf.id] = wf;
  });

  if (Object.keys(workflowIDToLatestOperators).length > 0) {
    return (
      <Box sx={{ display: 'flex', flexWrap: 'wrap' }}>
        {Object.entries(workflowIDToLatestOperators).map(
          ([wfId, operators]) => {
            return (
              <WorkflowSummaryCard
                integration={integration}
                key={wfId}
                workflow={workflowMap[wfId]}
                operators={operators}
              />
            );
          }
        )}
      </Box>
    );
  }

  return <Box>This resource is not used by any workflows.</Box>;
};

export default IntegrationWorkflowSummaryCards;
