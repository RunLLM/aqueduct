import { Link, Typography } from '@mui/material';
import Box from '@mui/material/Box';
import React, { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { handleFetchAllWorkflowSummaries } from '../../../reducers/listWorkflowSummaries';
import { AppDispatch, RootState } from '../../../stores/store';
import UserProfile from '../../../utils/auth';
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
import CheckItem from './components/CheckItem';
import ExecutionStatusLink from './components/ExecutionStatusLink';
import MetricItem from './components/MetricItem';
import ResourceItem from './components/ResourceItem';

type Props = {
  user: UserProfile;
  Layout?: React.FC<LayoutProps>;
};

const WorkflowsPage: React.FC<Props> = ({ user, Layout = DefaultLayout }) => {
  const dispatch: AppDispatch = useDispatch();

  useEffect(() => {
    document.title = 'Workflows | Aqueduct';
  }, []);

  useEffect(() => {
    dispatch(handleFetchAllWorkflowSummaries({ apiKey: user.apiKey }));
  }, [dispatch, user.apiKey]);

  const allWorkflows = useSelector(
    (state: RootState) => state.listWorkflowReducer
  );

  // If we are still loading the workflows, don't return a page at all.
  // Otherwise, we briefly return a page saying there are no workflows before
  // the workflows snap into place.
  if (
    allWorkflows.loadingStatus.loading === LoadingStatusEnum.Loading ||
    allWorkflows.loadingStatus.loading === LoadingStatusEnum.Initial
  ) {
    return null;
  }

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

  /**
   * Iterate through workflows array and map each element to a WorkflowTableRow object.
   */
  const workflowElements: PaginatedSearchTableRow[] = workflows.map((value) => {
    const engine = value.engine;

    let metrics = [];
    if (value?.metrics) {
      metrics = value.metrics.map((metric) => {
        return {
          metricId: metric.id,
          name: metric.name,
          value: metric.result?.content_serialized ?? '',
          status: metric.result?.exec_state?.status ?? ExecutionStatus.Unknown,
        };
      });
    }

    let checks = [];
    if (value.checks) {
      checks = value.checks.map((check) => {
        const value =
          check.result?.exec_state.status === 'succeeded' ? 'True' : 'False';

        return {
          checkId: check.id,
          name: check.name,
          level: check.spec?.check?.level ?? CheckLevel.Warning,
          timestamp: check.result?.exec_state?.timestamps?.finished_at ?? '',
          value,
          status: check.result?.exec_state?.status ?? ExecutionStatus.Unknown,
        };
      });
    }

    const workflowTableRow: PaginatedSearchTableRow = {
      name: {
        name: value.name,
        url: `/workflow/${value.id}`,
        status: value.status,
      },
      last_run: new Date(value.last_run_at * 1000),
      engine,
      metrics,
      checks,
    };

    return workflowTableRow;
  });

  const sortColumns = [
    {
      name: 'Last Run',
      sortAccessPath: ['last_run'],
    },
    {
      name: 'Name',
      sortAccessPath: ['name', 'name'],
    },
    {
      name: 'Engine',
      sortAccessPath: ['engine', 'engineName'],
    },
    {
      name: 'Status',
      sortAccessPath: ['name', 'status'],
    },
  ];

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
        value = row[column.name].toLocaleString();
        break;
      case 'engine': {
        value = <ResourceItem engine={value} />;
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

  const onChangeRowsPerPage = (rowsPerPage) => {
    localStorage.setItem('workflowsTableRowsPerPage', rowsPerPage);
  };

  const getRowsPerPage = () => {
    const savedRowsPerPage = localStorage.getItem('workflowsTableRowsPerPage');

    if (!savedRowsPerPage) {
      return 5; // return default rows per page value.
    }

    return parseInt(savedRowsPerPage);
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
          onChangeRowsPerPage={onChangeRowsPerPage}
          savedRowsPerPage={getRowsPerPage()}
          sortColumns={sortColumns}
        />
      ) : (
        <Box>{noItemsMessage}</Box>
      )}
    </Layout>
  );
};

export default WorkflowsPage;
