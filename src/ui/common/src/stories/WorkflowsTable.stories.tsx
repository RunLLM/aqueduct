import { Box } from '@mui/material';
import React from 'react';
import { CheckLevel } from '../utils/operators';

import WorkflowTable, {
  WorkflowTableData,
} from '../components/tables/WorkflowTable';
import { SupportedIntegrations } from '../utils/integrations';
import ExecutionStatus from '../utils/shared';
import CheckItem from '../components/pages/workflows/components/CheckItem';
import EngineItem from '../components/pages/workflows/components/EngineItem';
import MetricItem, { MetricPreview } from '../components/pages/workflows/components/MetricItem';
import WorkflowNameItem from '../components/pages/workflows/components/WorkflowNameItem';

export const WorkflowsTable: React.FC = () => {
  const checkPreviews = [
    {
      checkId: '1',
      name: 'min_churn',
      status: ExecutionStatus.Succeeded,
      level: CheckLevel.Error,
      value: 'True',
      timestamp: new Date().toLocaleString()
    },
    {
      checkId: '2',
      name: 'max_churn',
      status: ExecutionStatus.Failed,
      level: CheckLevel.Error,
      value: 'True',
      timestamp: new Date().toLocaleString()
    },
    {
      checkId: '3',
      name: 'avg_churn_check',
      // TODO: Come up with coherent color scheme for all of these different status levels.
      status: ExecutionStatus.Pending,
      level: CheckLevel.Warning,
      value: null,
      timestamp: new Date().toLocaleString()
    },
    {
      checkId: '4',
      name: 'warning_test',
      // TODO: Come up with coherent color scheme for all of these different status levels.
      status: ExecutionStatus.Succeeded,
      level: CheckLevel.Warning,
      value: 'False',
      timestamp: new Date().toLocaleString()
    },
  ];

  const checkTableItem = <CheckItem checks={checkPreviews} />;

  const metricsShort: MetricPreview[] = [
    { metricId: '1', name: 'avg_churn', value: '10' },
    { metricId: '2', name: 'sentiment', value: '100.5' },
    { metricId: '3', name: 'revenue_lost', value: '$20M' },
    { metricId: '4', name: 'more_metrics', value: '$500' },
  ];

  const metricsList = <MetricItem metrics={metricsShort} />;

  const airflowEngine = (
    <EngineItem
      engineName="airflow"
      engineIconUrl={SupportedIntegrations['Airflow'].logo}
    />
  );

  const lambdaEngine = (
    <EngineItem
      engineName="lambda"
      engineIconUrl={SupportedIntegrations['Lambda'].logo}
    />
  );

  const kubernetesEngine = (
    <EngineItem
      engineName="kubernetes"
      engineIconUrl={SupportedIntegrations['Kubernetes'].logo}
    />
  );

  const mockData: WorkflowTableData = {
    schema: {
      fields: [
        { name: 'name', type: 'varchar' },
        { name: 'last_run', type: 'varchar' },
        { name: 'engine', type: 'varchar' },
        { name: 'metrics', type: 'varchar' },
        { name: 'checks', type: 'varchar' },
      ],
      pandas_version: '1.5.1',
    },
    data: [
      {
        name: (
          <WorkflowNameItem name="churn" status={ExecutionStatus.Succeeded} />
        ),
        last_run: '11/1/2022 2:00PM',
        engine: airflowEngine,
        metrics: metricsList,
        checks: checkTableItem,
      },
      {
        name: (
          <WorkflowNameItem
            name="wine_ratings"
            status={ExecutionStatus.Failed}
          />
        ),
        last_run: '11/1/2022 2:00PM',
        engine: lambdaEngine,
        metrics: metricsList,
        checks: checkTableItem,
      },
      {
        name: (
          <WorkflowNameItem
            name="diabetes_classifier"
            status={ExecutionStatus.Pending}
          />
        ),
        last_run: '11/1/2022 2:00PM',
        engine: kubernetesEngine,
        metrics: metricsList,
        checks: checkTableItem,
      },
      {
        name: (
          <WorkflowNameItem
            name="mpg_regressor"
            status={ExecutionStatus.Canceled}
          />
        ),
        last_run: '11/1/2022 2:00PM',
        engine: lambdaEngine,
        metrics: metricsList,
        checks: checkTableItem,
      },
      {
        name: (
          <WorkflowNameItem
            name="house_price_prediction"
            status={ExecutionStatus.Registered}
          />
        ),
        last_run: '11/1/2022 2:00PM',
        engine: kubernetesEngine,
        metrics: metricsList,
        checks: checkTableItem,
      },
    ],
  };

  return (
    <Box>
      <WorkflowTable data={mockData} />
    </Box>
  );
};

export default WorkflowsTable;
