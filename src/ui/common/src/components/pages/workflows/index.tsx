import { Link, Typography } from '@mui/material';
import Box from '@mui/material/Box';
import React, { useEffect } from 'react';
import { useDispatch } from 'react-redux';

import {
  useDagGetQuery,
  useDagResultsGetQuery,
  useWorkflowsGetQuery,
} from '../../../handlers/AqueductApi';
import { AppDispatch } from '../../../stores/store';
import UserProfile from '../../../utils/auth';
import getPathPrefix from '../../../utils/getPathPrefix';
import ExecutionStatus from '../../../utils/shared';
import { getWorkflowEngineTypes } from '../../../utils/workflows';
import DefaultLayout from '../../layouts/default';
import { BreadcrumbLink } from '../../layouts/NavBar';
import {
  PaginatedSearchTable,
  SortType,
} from '../../tables/PaginatedSearchTable';
import { LayoutProps } from '../types';
import { useWorkflowNodes, useWorkflowNodesResults } from '../workflow/id/hook';
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

  const {
    data: workflowData,
    error: workflowError,
    isLoading: workflowLoading,
  } = useWorkflowsGetQuery({
    apiKey: user.apiKey,
  });

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

  const sortColumns = [
    {
      name: 'Last Run',
      sortAccessPath: ['Last Run', 'props', 'time'],
    },
    {
      name: 'Name',
      sortAccessPath: ['Name', 'props', 'name'],
    },
    {
      name: 'Engine',
      sortAccessPath: [
        'Engines',
        'props',
        'children',
        0,
        'props',
        'children',
        'props',
        'resource',
      ],
    },
    {
      name: 'Status',
      sortAccessPath: ['Name', 'props', 'status'],
    },
  ];

  const getLatestDagResult = (dagResults) =>
    dagResults.reduce(
      (prev, curr) =>
        curr.exec_state?.timestamps?.pending_at
          ? new Date(prev.exec_state?.timestamps?.pending_at) <
            new Date(curr.exec_state?.timestamps?.pending_at)
            ? curr
            : prev
          : curr,
      {
        exec_state: {
          status: ExecutionStatus.Registered,
          timestamps: { pending_at: 0 },
        },
      }
    );
  const LastRunComponent = (row) => {
    const workflowId = row.id;

    const {
      data: dagResults,
      error: dagResultsError,
      isLoading: dagResultsLoading,
    } = useDagResultsGetQuery({
      apiKey: user.apiKey,
      workflowId: workflowId,
    });

    let runTime = 'Not run yet.';
    let time = 0;

    if (!dagResultsLoading && !dagResultsError && dagResults.length > 0) {
      const latestDagResult = getLatestDagResult(dagResults);
      time = new Date(
        latestDagResult.exec_state?.timestamps?.pending_at
      ).getTime();
      runTime = new Date(
        latestDagResult.exec_state?.timestamps?.pending_at
      ).toLocaleString();
    }

    return <Typography time={time}>{runTime}</Typography>;
  };

  const columnAction = {
    Name: (row) => {
      const workflowId = row.id;
      const url = `${getPathPrefix()}/workflow/${workflowId}`;

      const {
        data: dagResults,
        error: dagResultsError,
        isLoading: dagResultsLoading,
      } = useDagResultsGetQuery({
        apiKey: user.apiKey,
        workflowId: workflowId,
      });
      let status = ExecutionStatus.Unknown;

      if (!dagResultsLoading && !dagResultsError && dagResults.length > 0) {
        const latestDagResult = getLatestDagResult(dagResults);
        if (latestDagResult) {
          status = latestDagResult.exec_state.status;
        }
      }

      return <ExecutionStatusLink name={row.name} url={url} status={status} />;
    },
    'Last Run': LastRunComponent,
    Engines: (row) => {
      const workflowId = row.id;

      const {
        data: dagResults,
        error: dagResultsError,
        isLoading: dagResultsLoading,
      } = useDagResultsGetQuery({
        apiKey: user.apiKey,
        workflowId: workflowId,
      });

      let latestDagId;
      if (!dagResultsLoading && !dagResultsError && dagResults.length > 0) {
        const latestDagResult = getLatestDagResult(dagResults);
        latestDagId = latestDagResult.dag_id;
      }
      const {
        data: dag,
        error: dagError,
        isLoading: dagLoading,
      } = useDagGetQuery(
        {
          apiKey: user.apiKey,
          workflowId: workflowId,
          dagId: latestDagId,
        },
        {
          skip: dagResultsLoading && latestDagId,
        }
      );

      const nodes = useWorkflowNodes(user.apiKey, workflowId, latestDagId);

      let engines = ['Unknown'];
      if (!dagLoading && !dagError && dag) {
        const workflowDag = structuredClone(dag);
        workflowDag.operators = nodes.operators;
        engines = getWorkflowEngineTypes(workflowDag);
      }

      return (
        <Box>
          {engines.map((v, idx) => (
            <Box
              mb={engines.length > 1 && idx < engines.length - 1 ? 1 : 0}
              key={idx}
            >
              {/* We need a box with margins so the chips have space between them. */}
              <ResourceItem resource={v} />
            </Box>
          ))}
        </Box>
      );
    },
    Metrics: (row) => {
      const workflowId = row.id;

      const {
        data: dagResults,
        error: dagResultsError,
        isLoading: dagResultsLoading,
      } = useDagResultsGetQuery({
        apiKey: user.apiKey,
        workflowId: workflowId,
      });

      let latestDagResultId;
      let latestDagId;
      if (!dagResultsLoading && !dagResultsError && dagResults.length > 0) {
        const latestDagResult = getLatestDagResult(dagResults);
        latestDagResultId = latestDagResult.id;
        latestDagId = latestDagResult.dag_id;
      }

      const nodes = useWorkflowNodes(user.apiKey, workflowId, latestDagId);
      const nodesResults = useWorkflowNodesResults(
        user.apiKey,
        workflowId,
        latestDagResultId
      );

      const metricNodes = Object.values(nodes.operators)
        .filter((op) => op.spec.type === 'metric')
        .map((op) => {
          const artifactId = op.outputs[0]; // Assuming there is only one output artifact
          return {
            metricId: op.id,
            name: op.name,
            value: nodesResults.artifacts[artifactId]?.content_serialized,
            status: nodesResults.artifacts[artifactId]?.exec_state?.status,
          };
        });
      return <MetricItem metrics={metricNodes} />;
    },
    Checks: (row) => {
      const workflowId = row.id;

      const {
        data: dagResults,
        error: dagResultsError,
        isLoading: dagResultsLoading,
      } = useDagResultsGetQuery({
        apiKey: user.apiKey,
        workflowId: workflowId,
      });

      let latestDagResultId;
      let latestDagId;
      if (!dagResultsLoading && !dagResultsError && dagResults.length > 0) {
        const latestDagResult = getLatestDagResult(dagResults);
        latestDagResultId = latestDagResult.id;
        latestDagId = latestDagResult.dag_id;
      }

      const nodes = useWorkflowNodes(user.apiKey, workflowId, latestDagId);
      const nodesResults = useWorkflowNodesResults(
        user.apiKey,
        workflowId,
        latestDagResultId
      );

      const checkNodes = Object.values(nodes.operators)
        .filter((op) => op.spec.type === 'check')
        .map((op) => {
          const artifactId = op.outputs[0]; // Assuming there is only one output artifact
          return {
            checkId: op.id,
            name: op.name,
            status: nodesResults.artifacts[artifactId]?.exec_state?.status,
            level: op.spec.check.level,
            value: nodesResults.artifacts[artifactId]?.content_serialized,
            timestamp:
              nodesResults.artifacts[artifactId]?.exec_state?.timestamps
                ?.finished_at,
          };
        });
      return <CheckItem checks={checkNodes} />;
    },
  };
  const columns = Object.keys(columnAction);

  const onGetColumnValue = (row, column) => {
    return columnAction[column](row);
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
      {workflowData && workflowData.length > 0 ? (
        <PaginatedSearchTable
          data={workflowData}
          columns={columns}
          searchEnabled={true}
          onGetColumnValue={onGetColumnValue}
          onChangeRowsPerPage={onChangeRowsPerPage}
          savedRowsPerPage={getRowsPerPage()}
          sortColumns={sortColumns}
          defaultSortConfig={{
            sortColumn: sortColumns[0],
            sortType: SortType.Descending,
          }}
        />
      ) : (
        <Box>{noItemsMessage}</Box>
      )}
    </Layout>
  );
};

export default WorkflowsPage;
