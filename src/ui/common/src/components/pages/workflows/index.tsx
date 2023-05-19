import { Link, Typography } from '@mui/material';
import Box from '@mui/material/Box';
import React, { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { handleFetchAllWorkflowSummaries } from '../../../reducers/listWorkflowSummaries';
import { AppDispatch, RootState } from '../../../stores/store';
import UserProfile from '../../../utils/auth';
import { CheckLevel } from '../../../utils/operators';
import ExecutionStatus, { LoadingStatusEnum } from '../../../utils/shared';
import { DagResultResponse, WorkflowResponse } from '../../../handlers/responses/Workflow';
import { reduceEngineTypes } from '../../../utils/workflows';
import DefaultLayout from '../../layouts/default';
import { BreadcrumbLink } from '../../layouts/NavBar';
import {
  PaginatedSearchTable,
  PaginatedSearchTableData,
  PaginatedSearchTableRow,
  SortType,
} from '../../tables/PaginatedSearchTable';
import { LayoutProps } from '../types';
import CheckItem from './components/CheckItem';
import ExecutionStatusLink from './components/ExecutionStatusLink';
import MetricItem from './components/MetricItem';
import ResourceItem from './components/ResourceItem';
import { useWorkflowsGetQuery, useDagResultsGetQuery, useNodesGetQuery, useNodesResultsGetQuery } from '../../../handlers/AqueductApi';
import getPathPrefix from '../../../utils/getPathPrefix';

type Props = {
  user: UserProfile;
  Layout?: React.FC<LayoutProps>;
};

const WorkflowsPage: React.FC<Props> = ({ user, Layout = DefaultLayout }) => {
  const dispatch: AppDispatch = useDispatch();

  useEffect(() => {
    document.title = 'Workflows | Aqueduct';
  }, []);

  const { data: workflowData, error: workflowError, isLoading: workflowLoading } = useWorkflowsGetQuery(
    {
      apiKey: user.apiKey,
    }
  );

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
 
  // TODO EUNICE: Refactor
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
      sortAccessPath: ['Engines', 'props', 'children', 0, 'props', 'children', 'props', 'resource'],
    },
    {
      name: 'Status',
      sortAccessPath: ['Name', 'props', 'status'],
    },
  ];

  const getLatestDagResult = (dagResults) => dagResults.reduce((prev, curr) => curr.exec_state?.timestamps?.pending_at? (new Date(prev.exec_state?.timestamps?.pending_at) < new Date(curr.exec_state?.timestamps?.pending_at)? curr : prev) : curr, {exec_state:{status:ExecutionStatus.Registered, timestamps:{pending_at:0}}});

  const columnAction = {
    "Name": (row) => {
      const workflowId = row.id;
      const url = `${getPathPrefix()}/workflow/${workflowId}`;

      const { data: dagResults, error: dagResultsError, isLoading: dagResultsLoading } = useDagResultsGetQuery(
        {
          apiKey: user.apiKey,
          workflowId: workflowId
        }
      );
      var status = ExecutionStatus.Unknown;

      if (!dagResultsLoading && !dagResultsError) {
        const latestDagResult = getLatestDagResult(dagResults);
        if (latestDagResult) {
          status = latestDagResult.exec_state.status;
        }
      }
      
      return <ExecutionStatusLink name={row.name} url={url} status={status} />;
    },
    "Last Run": (row) => {
      const workflowId = row.id;

      const { data: dagResults, error: dagResultsError, isLoading: dagResultsLoading } = useDagResultsGetQuery(
        {
          apiKey: user.apiKey,
          workflowId: workflowId
        }
      );

      var runTime = "Not run yet.";
      var time = 0;

      if (!dagResultsLoading && !dagResultsError && dagResults.length > 0) {
        const latestDagResult = getLatestDagResult(dagResults);
        time = new Date(latestDagResult.exec_state?.timestamps?.pending_at).getTime();
        runTime = new Date(latestDagResult.exec_state?.timestamps?.pending_at).toLocaleString();
      }

      return <Typography time={time}>{runTime}</Typography>; 
    },
    "Engines": (row) => {
      var engines = ["Unknown"];
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
      // const workflowId = row.id;

      // const { data: dagResults, error: dagResultsError, isLoading: dagResultsLoading } = useDagResultsGetQuery(
      //   {
      //     apiKey: user.apiKey,
      //     workflowId: workflowId
      //   }
      // );

      // if (!dagResultsLoading && !dagResultsError) {
      //   const latestDagResult = getLatestDagResult(dagResults);
      //   const latestDagId = latestDagResult.dag_id;

      //   getWorkflowEngineTypes(latestDag, op_eng_configs)
      // }
    
      // var engine = "Unknown";
      // getWorkflowEngineTypes
      // return engine;
    },
    "Metrics": (row) => {
      const workflowId = row.id;

      const { data: dagResults, error: dagResultsError, isLoading: dagResultsLoading } = useDagResultsGetQuery(
        {
          apiKey: user.apiKey,
          workflowId: workflowId
        }
      );

      var latestDagId;
      if (!dagResultsLoading && !dagResultsError) {
        const latestDagResult = getLatestDagResult(dagResults);
        latestDagId = latestDagResult.dag_id;
      }

      const { data: nodes, error: nodesError, isLoading: nodesLoading } = useNodesGetQuery(
        {
          apiKey: user.apiKey,
          dagId: latestDagId,
          workflowId: workflowId
        },
        {
          skip: dagResultsLoading,
        }
      );
      var metricNodes = [];
      if (!nodesLoading && !nodesError && nodes) {
        // Refactor once the metrics/checks updates are in because then it will be much easier.
        metricNodes = nodes.operators.filter((op) => op.spec.type === "metric").map((op) => {
          return {
            metricId: op.id,
            name: op.name,
          }
        })
      }
      return <MetricItem metrics={metricNodes} />; 
    },
    "Checks": (row) => {
      const workflowId = row.id;

      const { data: dagResults, error: dagResultsError, isLoading: dagResultsLoading } = useDagResultsGetQuery(
        {
          apiKey: user.apiKey,
          workflowId: workflowId
        }
      );

      var latestDagResultId;
      var latestDagId;
      if (!dagResultsLoading && !dagResultsError) {
        const latestDagResult = getLatestDagResult(dagResults);
        latestDagResultId = latestDagResult.id;
        latestDagId = latestDagResult.dag_id;
      }

      const { data: nodes, error: nodesError, isLoading: nodesLoading } = useNodesGetQuery(
        {
          apiKey: user.apiKey,
          dagId: latestDagId,
          workflowId: workflowId
        },
        {
          skip: dagResultsLoading,
        }
      );

      var checkNodes = [];
      if (!nodesLoading && !nodesError && nodes) {
        // Refactor once the metrics/checks updates are in because then it will be much easier.
        checkNodes = nodes.operators.filter((op) => op.spec.type === "check").map((op) => {
          const artifactId = op.outputs[0] // Assuming there is only one output artifact
          return {
            checkId: op.id,
            apiKey: user.apiKey,
            workflowId: workflowId,
            dagId: latestDagId,
            artifactId: artifactId,
            name: op.name,
            level: op.spec.check.level,
          }
        })
      }
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
