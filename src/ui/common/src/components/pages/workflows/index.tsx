import { Link, Typography } from '@mui/material';
import Box from '@mui/material/Box';
import React, { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { handleFetchAllWorkflowSummaries } from '../../../reducers/listWorkflowSummaries';
import { AppDispatch, RootState } from '../../../stores/store';
import UserProfile from '../../../utils/auth';
import { CheckLevel } from '../../../utils/operators';
import ExecutionStatus, { LoadingStatusEnum } from '../../../utils/shared';
import { WorkflowResponse } from '../../../handlers/responses/Workflow';
import { reduceEngineTypes } from '../../../utils/workflows';
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
import { useWorkflowsGetQuery, useDagResultsGetQuery } from '../../../handlers/AqueductApi';

type Props = {
  user: UserProfile;
  Layout?: React.FC<LayoutProps>;
};

const WorkflowsPage: React.FC<Props> = ({ user, Layout = DefaultLayout }) => {
  const dispatch: AppDispatch = useDispatch();

  useEffect(() => {
    document.title = 'Workflows | Aqueduct';
  }, []);

  let workflowData = null;

  const { data, error: workflowError, isLoading: workflowLoading } = useWorkflowsGetQuery(
    {
      apiKey: user.apiKey,
    },
    { 
      pollingInterval: 5000,
      skip: workflowData !== null,
    }
  );
  workflowData = data;

  console.log(workflowData, workflowError, workflowLoading);
// [
//     {
//         "id": "6cdea908-c3a7-401e-9081-157fc631e9f4",
//         "user_id": "3396169c-25cf-4613-9a2c-e7ac91edc4d5",
//         "name": "hotel_reviews",
//         "description": "",
//         "schedule": {
//             "trigger": "manual",
//             "cron_schedule": "",
//             "disable_manual_trigger": false,
//             "paused": false,
//             "source_id": "00000000-0000-0000-0000-000000000000"
//         },
//         "created_at": "2023-05-01T10:35:57.354686-07:00",
//         "retention_policy": {
//             "k_latest_runs": -1
//         },
//         "notification_settings": {
//             "settings": null
//         }
//     },
// ]
  
  // useEffect(() => {
  //   dispatch(handleFetchAllWorkflowSummaries({ apiKey: user.apiKey }));
  // }, [dispatch, user.apiKey]);
  
  // const allWorkflows = useSelector(
  //   (state: RootState) => state.listWorkflowReducer
  // );

  // workflowTableData = <Element workflowids={workflowData}/>

  // If we are still loading the workflows, don't return a page at all.
  // Otherwise, we briefly return a page saying there are no workflows before
  // the workflows snap into place.
  if (workflowLoading) {
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

  function getLatestDag(id: string) {
      const { data, error, isLoading } = useDagResultsGetQuery(
        {
          apiKey: user.apiKey,
          workflowId: id,
        },
        { pollingInterval: 5000 }
      );
      console.log("getLatestDag", id, data, error, isLoading); 
      if (isLoading) {
        return {
          id
        }
      } else {
        return {
          id
        }
      }
  }

  const workflows = data.map((workflow: WorkflowResponse) => getLatestDag(workflow.id));
  console.log("workflows", workflows);

  /**
   * Iterate through workflows array and map each element to a WorkflowTableRow object.
   */
  const workflowElements: PaginatedSearchTableRow[] = data.map(
    (workflow: WorkflowResponse) => {
      // const engines = reduceEngineTypes(
      //   workflow.engine,
      //   workflow.operator_engines.map((x) => (x ? x : workflow.engine))
      // );

      // let metrics = [];
      // if (workflow?.metrics) {
      //   metrics = workflow.metrics.map((metric) => {
      //     return {
      //       metricId: metric.id,
      //       name: metric.name,
      //       value: metric.result?.content_serialized ?? '',
      //       status:
      //         metric.result?.exec_state?.status ?? ExecutionStatus.Unknown,
      //     };
      //   });
      // }

      // let containsWarning = false;
      // let checks = [];
      // if (workflow.checks) {
      //   checks = workflow.checks.map((check) => {
      //     const value =
      //       check.result?.exec_state.status === 'succeeded' ? 'True' : 'False';
      //     const level = check.spec?.check?.level ?? CheckLevel.Warning;
      //     const status =
      //       check.result?.exec_state?.status ?? ExecutionStatus.Unknown;

      //     if (
      //       status === ExecutionStatus.Failed &&
      //       level === CheckLevel.Warning
      //     ) {
      //       containsWarning = true;
      //     }

      //     return {
      //       checkId: check.id,
      //       name: check.name,
      //       level,
      //       timestamp: check.result?.exec_state?.timestamps?.finished_at ?? '',
      //       value,
      //       status,
      //     };
      //   });
      // }
      // - latest dag status + last run + engines + checks / metrics results 
      const workflowTableRow: PaginatedSearchTableRow = {
        name: {
          name: workflow.name,
          url: `/workflow/${workflow.id}`,
          // Show warning badge if there is a warning check
          // status: containsWarning ? ExecutionStatus.Warning : workflow.status,
        },
        // last_run: new Date(workflow.last_run_at * 1000),
        // engines,
        // metrics,
        // checks,
      };

      return workflowTableRow;
    }
  );

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
      sortAccessPath: ['engines', 'engineName'],
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
        { name: 'engines', type: 'varchar' },
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
      case 'engines': {
        value = (
          <Box>
            {value.map((v, idx) => (
              <Box
                mb={value.length > 1 && idx < value.length - 1 ? 1 : 0}
                key={idx}
              >
                {/* We need a box with margins so the chips have space between them. */}
                <ResourceItem resource={v} />
              </Box>
            ))}
          </Box>
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
