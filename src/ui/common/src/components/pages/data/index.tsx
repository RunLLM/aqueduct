import { CircularProgress, Link, Typography } from '@mui/material';
import Box from '@mui/material/Box';
import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { BreadcrumbLink } from '../../../components/layouts/NavBar';
import { getDataArtifactPreview } from '../../../reducers/dataPreview';
import { handleLoadResources } from '../../../reducers/resources';
import { AppDispatch, RootState } from '../../../stores/store';
import UserProfile from '../../../utils/auth';
import getPathPrefix from '../../../utils/getPathPrefix';
import { CheckLevel } from '../../../utils/operators';
import ExecutionStatus from '../../../utils/shared';
import DefaultLayout from '../../layouts/default';
import PaginatedSearchTable, {
  PaginatedSearchTableData,
} from '../../tables/PaginatedSearchTable';
import { LayoutProps } from '../types';
import CheckItem from '../workflows/components/CheckItem';
import ExecutionStatusLink from '../workflows/components/ExecutionStatusLink';
import MetricItem from '../workflows/components/MetricItem';

type Props = {
  user: UserProfile;
  Layout?: React.FC<LayoutProps>;
};

const DataPage: React.FC<Props> = ({ user, Layout = DefaultLayout }) => {
  const apiKey = user.apiKey;
  const dispatch: AppDispatch = useDispatch();
  const [isLoading, setIsLoading] = useState<boolean>(false);

  useEffect(() => {
    const fetchDataArtifactsAndResources = async () => {
      setIsLoading(true);
      await dispatch(getDataArtifactPreview({ apiKey }));
      await dispatch(handleLoadResources({ apiKey }));
      setIsLoading(false);
    };

    fetchDataArtifactsAndResources();
  }, [apiKey, dispatch]);

  const dataCardsInfo = useSelector(
    (state: RootState) => state.dataPreviewReducer
  );

  useEffect(() => {
    document.title = 'Data | Aqueduct';
  }, []);

  const onGetColumnValue = (row, column) => {
    let value = row[column.name];

    switch (column.name) {
      case 'workflow':
      case 'name':
        const { name, url, status } = value;
        value = <ExecutionStatusLink name={name} url={url} status={status} />;
        break;
      case 'created_at':
        value = row[column.name].toLocaleString();
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

  let tableData = [];
  if (dataCardsInfo?.data?.latest_versions) {
    const latestVersions =
      Object.keys(dataCardsInfo.data.latest_versions).length > 0
        ? Object.keys(dataCardsInfo.data.latest_versions)
        : [];

    tableData = latestVersions.map((version) => {
      const currentVersion =
        dataCardsInfo.data.latest_versions[version.toString()];

      const artifactId = currentVersion.artifact_id;
      const artifactName = currentVersion.artifact_name;

      const dataPreviewInfoVersions = Object.entries(currentVersion.versions);
      let [latestDagResultId, latestVersion] =
        dataPreviewInfoVersions.length > 0 ? dataPreviewInfoVersions[0] : null;

      // Find the latest version
      // note: could also sort the array and get things that way.
      dataPreviewInfoVersions.forEach(([dagResultId, version]) => {
        if (version.timestamp > latestVersion.timestamp) {
          latestDagResultId = dagResultId;
          latestVersion = version;
        }
      });

      let checks = [];
      if (latestVersion?.checks?.length > 0) {
        checks = latestVersion?.checks.map((check, index) => {
          const level = check.metadata.failure_type
            ? CheckLevel.Warning
            : CheckLevel.Error;
          const value =
            check.metadata.status === 'succeeded' && !check.metadata.error;
          return {
            checkId: index,
            name: check.name,
            status: check.status,
            level,
            value: value ? 'True' : 'False',
            timestamp: check.metadata.timestamps.finished_at,
          };
        });
      }

      let metrics = [];
      if (latestVersion?.metrics?.length > 0) {
        metrics = latestVersion?.metrics?.map((metric) => {
          return {
            metricId: metric.id,
            name: metric.name,
            value: metric.result.content_serialized,
            status:
              metric.result?.exec_state?.status ?? ExecutionStatus.Unknown,
          };
        });
      }

      const workflowId = currentVersion.workflow_id;
      const workflowName = currentVersion.workflow_name;

      return {
        name: {
          name: artifactName,
          url: `${getPathPrefix()}/workflow/${workflowId}/result/${latestDagResultId}/artifact/${artifactId}`,
          status: latestVersion.status,
        },
        created_at: new Date(latestVersion.timestamp * 1000),
        workflow: {
          name: workflowName,
          url: `${getPathPrefix()}/workflow/${workflowId}`,
          status: latestVersion?.dag_status ?? ExecutionStatus.Unknown,
        },
        type: latestVersion?.metadata?.python_type ?? '-',
        metrics,
        checks,
      };
    });
  }

  const sortColumns = [
    {
      name: 'Name',
      sortAccessPath: ['name', 'name'],
    },
    {
      name: 'Created At',
      sortAccessPath: ['created_at'],
    },
    {
      name: 'Type',
      sortAccessPath: ['type'],
    },
    {
      name: 'Status',
      sortAccessPath: ['name', 'status'],
    },
  ];

  const artifactList: PaginatedSearchTableData = {
    schema: {
      fields: [
        { name: 'name', type: 'varchar' },
        { name: 'created_at', displayName: 'Created At', type: 'varchar' },
        { name: 'workflow', type: 'varchar' },
        { name: 'type', type: 'varchar' },
        { name: 'metrics', type: 'varchar' },
        { name: 'checks', type: 'varchar' },
      ],
      pandas_version: '1.5.1',
    },
    data: tableData,
  };

  const noItemsMessage = (
    <Typography variant="h5">
      There are no data artifacts created yet. Create one right by running a
      workflow with our{' '}
      <Link href="https://github.com/aqueducthq/aqueduct/blob/main/sdk">
        Python SDK
      </Link>
      <span>!</span>
    </Typography>
  );

  const onChangeRowsPerPage = (rowsPerPage) => {
    localStorage.setItem('dataTableRowsPerPage', rowsPerPage);
  };

  const getRowsPerPage = () => {
    const savedRowsPerPage = localStorage.getItem('dataTableRowsPerPage');

    if (!savedRowsPerPage) {
      return 5; // return default rows per page value.
    }

    return parseInt(savedRowsPerPage);
  };

  if (isLoading) {
    return (
      <Layout
        breadcrumbs={[BreadcrumbLink.HOME, BreadcrumbLink.DATA]}
        user={user}
      >
        <Box
          sx={{
            display: 'flex',
            width: '100%',
            height: '100%',
            justifyContent: 'center',
            alignItems: 'center',
          }}
        >
          <Box sx={{ width: '64px', height: '64px' }}>
            <CircularProgress sx={{ width: '100%', height: '100%' }} />
          </Box>
        </Box>
      </Layout>
    );
  }

  return (
    <Layout
      breadcrumbs={[BreadcrumbLink.HOME, BreadcrumbLink.DATA]}
      user={user}
    >
      {artifactList.data?.length && artifactList.data?.length > 0 ? (
        <PaginatedSearchTable
          data={artifactList}
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

export default DataPage;
