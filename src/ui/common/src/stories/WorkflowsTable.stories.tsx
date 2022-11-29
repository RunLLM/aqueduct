import { Box } from '@mui/material';
import React from 'react';

import CheckItem, {
  CheckPreview,
} from '../components/pages/workflows/components/CheckItem';
import EngineItem from '../components/pages/workflows/components/EngineItem';
import ExecutionStatusLink from '../components/pages/workflows/components/ExecutionStatusLink';
import MetricItem, {
  MetricPreview,
} from '../components/pages/workflows/components/MetricItem';
import WorkflowTable, {
  WorkflowTableData,
} from '../components/tables/WorkflowTable';
import { SupportedIntegrations } from '../utils/integrations';
import { CheckLevel } from '../utils/operators';
import ExecutionStatus from '../utils/shared';

export const WorkflowsTable: React.FC = () => {
  const checkPreviews: CheckPreview[] = [
    {
      checkId: '1',
      name: 'min_churn',
      status: ExecutionStatus.Succeeded,
      level: CheckLevel.Error,
      value: 'True',
      timestamp: new Date().toLocaleString(),
    },
    {
      checkId: '2',
      name: 'max_churn',
      status: ExecutionStatus.Failed,
      level: CheckLevel.Error,
      value: 'True',
      timestamp: new Date().toLocaleString(),
    },
    {
      checkId: '3',
      name: 'avg_churn_check',
      // TODO: Come up with coherent color scheme for all of these different status levels.
      status: ExecutionStatus.Pending,
      level: CheckLevel.Warning,
      value: null,
      timestamp: new Date().toLocaleString(),
    },
    {
      checkId: '4',
      name: 'warning_test',
      // TODO: Come up with coherent color scheme for all of these different status levels.
      status: ExecutionStatus.Succeeded,
      level: CheckLevel.Warning,
      value: 'False',
      timestamp: new Date().toLocaleString(),
    },
    {
      checkId: '5',
      name: 'canceled_test',
      // TODO: Come up with coherent color scheme for all of these different status levels.
      status: ExecutionStatus.Canceled,
      level: CheckLevel.Warning,
      value: 'False',
      timestamp: new Date().toLocaleString(),
    },
  ];

  const metricsShort: MetricPreview[] = [
    {
      metricId: '1',
      name: 'avg_churn',
      value: '10',
      status: ExecutionStatus.Failed,
    },
    {
      metricId: '2',
      name: 'sentiment',
      value: '100.5',
      status: ExecutionStatus.Succeeded,
    },
    {
      metricId: '3',
      name: 'revenue_lost',
      value: '$20M',
      status: ExecutionStatus.Succeeded,
    },
    {
      metricId: '4',
      name: 'more_metrics',
      value: '$500',
      status: ExecutionStatus.Succeeded,
    },
  ];

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
        name: {
          name: 'churn',
          url: '/workflows',
          status: ExecutionStatus.Succeeded,
        },
        last_run: '11/1/2022 2:00PM',
        //engine: airflowEngine,
        engine: {
          engineName: 'kubernetes',
          engineIconUrl: SupportedIntegrations['Kubernetes'].logo,
        },
        metrics: metricsShort,
        checks: checkPreviews,
      },
      {
        name: {
          name: 'wine_ratings',
          url: '/workflows',
          status: ExecutionStatus.Succeeded,
        },
        last_run: '11/1/2022 2:00PM',
        //engine: lambdaEngine,
        engine: {
          engineName: 'lambda',
          engineIconUrl: SupportedIntegrations['Lambda'].logo,
        },
        metrics: metricsShort,
        checks: checkPreviews,
      },
      {
        name: {
          name: 'diabetes_classifier',
          url: '/workflows',
          status: ExecutionStatus.Pending,
        },
        last_run: '11/1/2022 2:00PM',
        engine: {
          engineName: 'kubernetes',
          engineIconUrl: SupportedIntegrations['Kubernetes'].logo,
        },
        metrics: metricsShort,
        checks: checkPreviews,
      },
      {
        name: {
          name: 'mpg_regressor',
          url: '/workflows',
          status: ExecutionStatus.Canceled,
        },
        last_run: '11/1/2022 2:00PM',
        engine: {
          engineName: 'lambda',
          engineIconUrl: SupportedIntegrations['Lambda'].logo,
        },
        metrics: metricsShort,
        checks: checkPreviews,
      },
      {
        name: {
          name: 'house_price_prediction',
          url: '/workflows',
          status: ExecutionStatus.Registered,
        },
        last_run: '11/1/2022 2:00PM',
        engine: {
          engineName: 'kubernetes',
          engineIconUrl: SupportedIntegrations['Kubernetes'].logo,
        },
        metrics: metricsShort,
        checks: checkPreviews,
      },
    ],
    meta: [
      {
        name: 'churn',
        last_run: '11/1/2022 2:00PM',
        engine: 'airflow',
        metrics: metricsShort,
        checks: checkPreviews,
      },
      {
        name: 'wine_ratings',
        last_run: '11/1/2022 2:00PM',
        engine: 'lambda',
        metrics: metricsShort,
        checks: checkPreviews,
      },
      {
        name: 'diabetes_classifier',
        last_run: '11/1/2022 2:00PM',
        engine: 'kubernetes',
        metrics: metricsShort,
        checks: checkPreviews,
      },
      {
        name: 'mpg_regressor',
        last_run: '11/1/2022 2:00PM',
        engine: 'lambda',
        metrics: metricsShort,
        checks: checkPreviews,
      },
      {
        name: 'house_price_prediction',
        last_run: '11/1/2022 2:00PM',
        engine: 'kubernetes',
        metrics: metricsShort,
        checks: checkPreviews,
      },
    ],
  };

  const onGetColumnValue = (row, column) => {
    let value = row[column.name];

    switch (column.name) {
      case 'name':
        const { name, url, status } = value;
        value = <ExecutionStatusLink name={name} url={url} status={status} />;
        break;
      case 'last_run':
        value = row[column.name];
        break;
      case 'engine': {
        const { engineName, engineIconUrl } = value;
        value = (
          <EngineItem engineName={engineName} engineIconUrl={engineIconUrl} />
        );
        break;
      }
      case 'metrics': {
        value = <MetricItem metrics={value} />;
        break;
      }
      case 'checks': {
        value = <CheckItem checks={value} />;
        break;
      }
      default: {
        value = row[column.name];
        break;
      }
    }

    return value;
  };

  return (
    <Box>
      <WorkflowTable
        data={mockData}
        searchEnabled={true}
        onGetColumnValue={onGetColumnValue}
      />
    </Box>
  );
};

export default WorkflowsTable;
