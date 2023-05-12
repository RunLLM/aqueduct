import { Box } from '@mui/material';
import React from 'react';

import {
  CheckItem,
  CheckPreview,
} from '../components/pages/workflows/components/CheckItem';
import ExecutionStatusLink from '../components/pages/workflows/components/ExecutionStatusLink';
import ResourceItem from '../components/pages/workflows/components/ResourceItem';
import PaginatedSearchTable, {
  PaginatedSearchTableData,
} from '../components/tables/PaginatedSearchTable';
import { ServiceLogos } from '../utils/integrations';
import { CheckLevel } from '../utils/operators';
import ExecutionStatus from '../utils/shared';
import { ComponentMeta } from '@storybook/react';
import { MetricItem, MetricPreview } from '../components/pages/workflows/components/MetricItem';

const WorkflowsTable: React.FC = () => {
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
      status: ExecutionStatus.Pending,
      level: CheckLevel.Warning,
      value: null,
      timestamp: new Date().toLocaleString(),
    },
    {
      checkId: '4',
      name: 'warning_test',
      status: ExecutionStatus.Succeeded,
      level: CheckLevel.Warning,
      value: 'False',
      timestamp: new Date().toLocaleString(),
    },
    {
      checkId: '5',
      name: 'canceled_test',
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

  const mockData: PaginatedSearchTableData = {
    schema: {
      fields: [
        { name: 'name', type: 'varchar' },
        { name: 'last_run', displayName: 'Last Run', type: 'varchar' },
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
          engineIconUrl: ServiceLogos['Kubernetes'].logo,
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
        engine: {
          engineName: 'lambda',
          engineIconUrl: ServiceLogos['Lambda'].logo,
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
          engineIconUrl: ServiceLogos['Kubernetes'].logo,
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
          engineIconUrl: ServiceLogos['Lambda'].logo,
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
          engineIconUrl: ServiceLogos['Kubernetes'].logo,
        },
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
          <ResourceItem engineName={engineName} engineIconUrl={engineIconUrl} />
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
      <PaginatedSearchTable
        data={mockData}
        searchEnabled={true}
        onGetColumnValue={onGetColumnValue}
      />
    </Box>
  );
};

export default {
  title: 'Components/WorkflowsTable',
  component: WorkflowsTable,
  argTypes: {},
} as ComponentMeta<typeof WorkflowsTable>;