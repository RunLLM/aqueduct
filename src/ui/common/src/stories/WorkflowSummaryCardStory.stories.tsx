import { Box } from '@mui/material';
import React from 'react';

import WorkflowSummaryCard, {
  WorkflowSummaryCardProps,
} from '../components/workflows/WorkflowSummaryCard';
import { EngineType } from '../utils/engine';
import ExecutionStatus from '../utils/shared';

export const WorkflowSummaryCardStory: React.FC = () => {
  const workflowSummaries: WorkflowSummaryCardProps[] = [
    {
      workflow: {
        id: '1',
        name: 'Workflow 1',
        description: 'This is a workflow',
        created_at: Date.now() / 1000,
        last_run_at: Date.now() / 1000,
        status: ExecutionStatus.Succeeded,
        engine: EngineType.Airflow,
        operator_engines: [EngineType.Airflow],
        metrics: [],
        checks: [],
      },
      operators: [],
      integration: {
        id: '1',
        service: 'Postgres',
        name: 'Postgres Resource',
        createdAt: Date.now() / 1000,
        exec_state: {
          status: ExecutionStatus.Succeeded,
        },
        config: {
          host: 'aam19861.us-east-2.amazonaws.com',
          port: '5432',
          database: 'prod',
          username: 'prod-pg-aq',
        },
      },
    },
  ];

  return (
    <Box sx={{ display: 'flex', flexWrap: 'wrap', alignImems: 'flex-start' }}>
      {workflowSummaries.map((wf) => (
        <WorkflowSummaryCard key={wf.workflow.id} {...wf} />
      ))}
    </Box>
  );
};

export default WorkflowSummaryCardStory;
