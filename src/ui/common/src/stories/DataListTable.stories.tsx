import { Box, Typography } from '@mui/material';
import React from 'react';

import CheckItem, {
  CheckPreview,
} from '../components/pages/workflows/components/CheckItem';
import ExecutionStatusLink from '../components/pages/workflows/components/ExecutionStatusLink';
import MetricItem, {
  MetricPreview,
} from '../components/pages/workflows/components/MetricItem';
import WorkflowTable, {
  WorkflowTableData,
} from '../components/tables/WorkflowTable';
import { CheckLevel } from '../utils/operators';
import ExecutionStatus from '../utils/shared';

export const DataListTable: React.FC = () => {
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
      status: ExecutionStatus.Succeeded,
    },
    {
      metricId: '2',
      name: 'sentiment',
      value: '100.5',
      status: ExecutionStatus.Failed,
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
    {
      metricId: '5',
      name: 'more_metrics',
      value: '$500',
      status: ExecutionStatus.Succeeded,
    },
  ];

  // TODO: Change this type to something more generic.
  // Also make this change in WorkflowsTable, I think we can just use Data here if we add JSX.element to Data's union type.
  const mockData: WorkflowTableData = {
    schema: {
      fields: [
        { name: 'name', type: 'varchar' },
        { name: 'created_at', type: 'varchar' },
        { name: 'workflow', type: 'varchar' },
        { name: 'type', type: 'varchar' },
        { name: 'metrics', type: 'varchar' },
        { name: 'checks', type: 'varchar' },
      ],
      pandas_version: '1.5.1',
    },
    data: [
      {
        name: {
          name: 'churn_model',
          url: '/data',
          status: ExecutionStatus.Succeeded,
        },
        created_at: '11/1/2022 2:00PM',
        workflow: {
          name: 'train_churn_model',
          url: '/workflows',
          status: ExecutionStatus.Running,
        },
        type: 'sklearn.linear, Linear Regression',
        metrics: metricsShort,
        checks: checkPreviews,
      },
      {
        name: {
          name: 'predict_churn_dataset',
          url: '/workflows',
          status: ExecutionStatus.Running,
        },
        created_at: '11/1/2022 2:00PM',
        workflow: {
          name: 'monthly_churn_prediction',
          url: '/workflows',
          status: ExecutionStatus.Succeeded,
        },
        type: 'pandas.DataFrame',
        metrics: metricsShort,
        checks: checkPreviews,
      },
      {
        name: {
          name: 'label_classifier',
          url: '/data',
          status: ExecutionStatus.Pending,
        },
        created_at: '11/1/2022 2:00PM',
        workflow: {
          name: 'label_classifier_workflow',
          url: '/workflows',
          status: ExecutionStatus.Registered,
        },
        type: 'parquet',
        metrics: metricsShort,
        checks: checkPreviews,
      },
    ],
    meta: [
      {
        name: 'churn_model',
        created_at: '11/1/2022 2:00PM',
        workflow: 'train_churn_model',
        type: 'sklearn.linear, Linear Regression',
        metrics: metricsShort,
        checks: checkPreviews,
      },
      {
        name: 'predict_churn_dataset',
        created_at: '11/1/2022 2:00PM',
        workflow: 'monthly_churn_prediction',
        type: 'pandas.DataFrame',
        metrics: metricsShort,
        checks: checkPreviews,
      },
      {
        name: 'label_classifier',
        created_at: '11/1/2022 2:00PM',
        workflow: 'label_classifier_workflow',
        type: 'parquet',
        metrics: metricsShort,
        checks: checkPreviews,
      },
    ],
  };

  const onGetColumnValue = (row, column) => {
    let value = row[column.name];

    console.log('column: ', column.name);

    switch (column.name) {
      case 'workflow':
      case 'name':
        const { name, url, status } = value;
        value = <ExecutionStatusLink name={name} url={url} status={status} />;
        break;
      case 'created_at':
        value = row[column.name];
        break;
      case 'metrics': {
        value = <MetricItem metrics={value} />;
        break;
      }
      case 'checks': {
        value = <CheckItem checks={value} />;
        break;
      }
      case 'type': {
        console.log('inside dataTable onGetColumnValue');
        value = (
          <Typography fontFamily="monospace">{row[column.name]}</Typography>
        );
        break;
      }
      default: {
        value = row[column.name];
        break;
      }
    }

    return value;
  };

  // TODO: Rename "WorkflowTable" to something more generic.
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

export default DataListTable;
