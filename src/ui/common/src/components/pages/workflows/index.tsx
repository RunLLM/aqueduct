import { Link, Typography } from '@mui/material';
import Box from '@mui/material/Box';
import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { SupportedIntegrations } from '../../../utils/integrations';
import { WorkflowTableData, WorkflowTable, WorkflowTableRow } from '../../../components/tables/WorkflowTable';

import { handleFetchAllWorkflowSummaries } from '../../../reducers/listWorkflowSummaries';
import { AppDispatch, RootState } from '../../../stores/store';
import UserProfile from '../../../utils/auth';
import ExecutionStatus, { LoadingStatusEnum } from '../../../utils/shared';
import { ListWorkflowSummary } from '../../../utils/workflows';
import { CardPadding } from '../../layouts/card';
import DefaultLayout from '../../layouts/default';
import { BreadcrumbLink } from '../../layouts/NavBar';
import { filteredList, SearchBar } from '../../Search';
import WorkflowCard from '../../workflows/workflowCard';
import { LayoutProps } from '../types';
import { CheckLevel } from '../../../utils/operators';
import CheckItem, { CheckPreview } from './components/CheckItem';
import MetricItem, { MetricPreview } from './components/MetricItem';
import EngineItem from './components/EngineItem';
import ExecutionStatusLink from './components/ExecutionStatusLink';

type Props = {
  user: UserProfile;
  Layout?: React.FC<LayoutProps>;
};

const WorkflowsPage: React.FC<Props> = ({ user, Layout = DefaultLayout }) => {
  const dispatch: AppDispatch = useDispatch();
  const [filterText, setFilterText] = useState<string>('');

  useEffect(() => {
    document.title = 'Workflows | Aqueduct';
  }, []);

  useEffect(() => {
    dispatch(handleFetchAllWorkflowSummaries({ apiKey: user.apiKey }));
  }, [dispatch, user.apiKey]);

  const allWorkflows = useSelector(
    (state: RootState) => state.listWorkflowReducer
  );

  const getOptionLabel = (workflow) => workflow.name;

  // If we are still loading the workflows, don't return a page at all.
  // Otherwise, we briefly return a page saying there are no workflows before
  // the workflows snap into place.
  if (
    allWorkflows.loadingStatus.loading === LoadingStatusEnum.Loading ||
    allWorkflows.loadingStatus.loading === LoadingStatusEnum.Initial
  ) {
    return null;
  }

  const displayFilteredWorkflows = (workflow) => {
    return (
      <Box my={2}>
        <WorkflowCard workflow={workflow} />
      </Box>
    );
  };

  const noItemsMessage = (
    <Typography variant="h5">
      There are no workflows created yet. Create one right now with our{' '}
      <Link href="https://github.com/aqueducthq/aqueduct/blob/main/sdk">
        Python SDK
      </Link>
      <span>!</span>
    </Typography>
  );

  const workflowList = filteredList(
    filterText,
    allWorkflows.workflows,
    getOptionLabel,
    displayFilteredWorkflows,
    noItemsMessage
  );

  console.log('allWorkflows: ', allWorkflows);
  const workflows = allWorkflows.workflows;

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

  const workflowElements: WorkflowTableRow[] = workflows.map((value) => {
    console.log('engine: ', value.engine);
    const engineName = value.engine[0].toUpperCase() + value.engine.substring(1);
    console.log('engineName: ', engineName);

    const engineIconUrl = SupportedIntegrations[value.engine[0].toUpperCase() + value.engine.substring(1)].logo;
    console.log('iconUrl: ', engineIconUrl);

    const workflowTableRow: WorkflowTableRow = {
      name: {
        name: value.name,
        url: `/workflow/${value.id}`,
      },
      last_run: value.last_run_at,
      engine: {
        engineName: value.engine,
        engineIconUrl: SupportedIntegrations[value.engine[0].toUpperCase() + value.engine.substring(1)].logo
      },
      metrics: metricsShort,
      checks: checkPreviews
    };

    return workflowTableRow;
  });

  const workflowTableData: WorkflowTableData = {
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
    meta: [],
  };

  console.log('workflowTableData: ', workflowTableData);

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
      <Box>
        <WorkflowTable
          data={workflowTableData}
          searchEnabled={true}
          onGetColumnValue={onGetColumnValue}
        />
      </Box>
      <Box>
        {allWorkflows.workflows.length >= 1 && (
          <Box marginLeft={CardPadding}>
            {/* Align searchbar with card text */}
            <SearchBar
              options={allWorkflows.workflows}
              getOptionLabel={(option: ListWorkflowSummary) =>
                option.name || ''
              }
              setSearchTerm={setFilterText}
            />
          </Box>
        )}
        {workflowList}
      </Box>
    </Layout>
  );
};

export default WorkflowsPage;
