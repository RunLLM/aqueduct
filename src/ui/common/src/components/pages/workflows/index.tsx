import { Link, Typography } from '@mui/material';
import Box from '@mui/material/Box';
import React, { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { handleFetchAllWorkflowSummaries } from '../../../reducers/listWorkflowSummaries';
import { AppDispatch, RootState } from '../../../stores/store';
import UserProfile from '../../../utils/auth';
import { SupportedIntegrations } from '../../../utils/integrations';
import { CheckLevel } from '../../../utils/operators';
import ExecutionStatus, { LoadingStatusEnum } from '../../../utils/shared';
import DefaultLayout from '../../layouts/default';
import { BreadcrumbLink } from '../../layouts/NavBar';
import {
  PaginatedSearchTable,
  PaginatedSearchTableData,
  PaginatedSearchTableRow,
} from '../../tables/PaginatedSearchTable';
import { LayoutProps } from '../types';
import CheckItem, { CheckPreview } from './components/CheckItem';
import EngineItem from './components/EngineItem';
import ExecutionStatusLink from './components/ExecutionStatusLink';
import MetricItem, { MetricPreview } from './components/MetricItem';

type Props = {
  user: UserProfile;
  Layout?: React.FC<LayoutProps>;
};

const WorkflowsPage: React.FC<Props> = ({ user, Layout = DefaultLayout }) => {
  const dispatch: AppDispatch = useDispatch();
  //const [filterText, setFilterText] = useState<string>('');

  useEffect(() => {
    document.title = 'Workflows | Aqueduct';
  }, []);

  useEffect(() => {
    dispatch(handleFetchAllWorkflowSummaries({ apiKey: user.apiKey }));
  }, [dispatch, user.apiKey]);

  const allWorkflows = useSelector(
    (state: RootState) => state.listWorkflowReducer
  );

  //const getOptionLabel = (workflow) => workflow.name;

  // If we are still loading the workflows, don't return a page at all.
  // Otherwise, we briefly return a page saying there are no workflows before
  // the workflows snap into place.
  if (
    allWorkflows.loadingStatus.loading === LoadingStatusEnum.Loading ||
    allWorkflows.loadingStatus.loading === LoadingStatusEnum.Initial
  ) {
    return null;
  }

  // TODO: Deprecate this card and searchbar.
  // const displayFilteredWorkflows = (workflow) => {
  //   return (
  //     <Box my={2}>
  //       <WorkflowCard workflow={workflow} />
  //     </Box>
  //   );
  // };

  // const workflowList = filteredList(
  //   filterText,
  //   allWorkflows.workflows,
  //   getOptionLabel,
  //   displayFilteredWorkflows,
  //   noItemsMessage
  // );

  const noItemsMessage = (
    <Typography variant="h5">
      There are no workflows created yet. Create one right now with our{' '}
      <Link href="https://github.com/aqueducthq/aqueduct/blob/main/sdk">
        Python SDK
      </Link>
      <span>!</span>
    </Typography>
  );

  const workflows = allWorkflows.workflows;

  // TODO: Remove these once we fetch actual metrics data from the API.
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

  // TODO: Remove these once we fetch actual checks data from the API
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

  /**
   * Iterate through workflows array and map each element to a WorkflowTableRow object.
   */
  const workflowElements: PaginatedSearchTableRow[] = workflows.map((value) => {
    const engineName =
      value.engine[0].toUpperCase() + value.engine.substring(1);
    const engineIconUrl =
      SupportedIntegrations[
        value.engine[0].toUpperCase() + value.engine.substring(1)
      ].logo;
    const workflowTableRow: PaginatedSearchTableRow = {
      name: {
        name: value.name,
        url: `/workflow/${value.id}`,
        status: value.status,
      },
      // TODO: Figur out correct way to render this date string
      last_run: new Date(value.last_run_at * 1000).toLocaleString(),
      engine: {
        engineName,
        engineIconUrl: engineIconUrl,
      },
      // TODO: add metrics and checks to response item when getting workflows list.
      metrics: metricsShort,
      checks: checkPreviews,
    };

    return workflowTableRow;
  });

  const workflowTableData: PaginatedSearchTableData = {
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
    data: workflowElements,
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
    <Layout
      breadcrumbs={[BreadcrumbLink.HOME, BreadcrumbLink.WORKFLOWS]}
      user={user}
    >
      {workflowTableData.data.length > 0 ? (
        <PaginatedSearchTable
          data={workflowTableData}
          searchEnabled={true}
          onGetColumnValue={onGetColumnValue}
        />
      ) : (
        <Box>{noItemsMessage}</Box>
      )}
    </Layout>
  );
};

export default WorkflowsPage;
