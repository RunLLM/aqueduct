import { Typography } from '@mui/material';
import Box from '@mui/material/Box';
import React, { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { BreadcrumbLink } from '../../../components/layouts/NavBar';
import WorkflowTable, {
  WorkflowTableData,
} from '../../../components/tables/WorkflowTable';
import { getDataArtifactPreview } from '../../../reducers/dataPreview';
import { handleLoadIntegrations } from '../../../reducers/integrations';
import { AppDispatch, RootState } from '../../../stores/store';
import UserProfile from '../../../utils/auth';
import getPathPrefix from '../../../utils/getPathPrefix';
import { CheckLevel } from '../../../utils/operators';
import ExecutionStatus from '../../../utils/shared';
import DefaultLayout from '../../layouts/default';
import { LayoutProps } from '../types';
import CheckItem from '../workflows/components/CheckItem';
import ExecutionStatusLink from '../workflows/components/ExecutionStatusLink';
import MetricItem, { MetricPreview } from '../workflows/components/MetricItem';

type Props = {
  user: UserProfile;
  Layout?: React.FC<LayoutProps>;
};

const DataPage: React.FC<Props> = ({ user, Layout = DefaultLayout }) => {
  const apiKey = user.apiKey;
  const dispatch: AppDispatch = useDispatch();

  useEffect(() => {
    dispatch(getDataArtifactPreview({ apiKey }));
    dispatch(handleLoadIntegrations({ apiKey }));
  }, [apiKey, dispatch]);

  const dataCardsInfo = useSelector(
    (state: RootState) => state.dataPreviewReducer
  );

  useEffect(() => {
    document.title = 'Data | Aqueduct';
  }, []);

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

  const onGetColumnValue = (row, column) => {
    let value = row[column.name];

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

  if (Object.keys(dataCardsInfo.data.latest_versions).length > 0) {
    tableData = Object.keys(dataCardsInfo.data.latest_versions).map(
      (version) => {
        const currentVersion =
          dataCardsInfo.data.latest_versions[version.toString()];

        const artifactId = currentVersion.artifact_id;
        const artifactName = currentVersion.artifact_name;

        const dataPreviewInfoVersions = Object.entries(currentVersion.versions);
        let [latestDagResultId, latestVersion] =
          dataPreviewInfoVersions.length > 0
            ? dataPreviewInfoVersions[0]
            : null;

        // Find the latest version
        // note: could also sort the array and get things that way.
        dataPreviewInfoVersions.forEach(([dagResultId, version]) => {
          if (version.timestamp > latestVersion.timestamp) {
            latestDagResultId = dagResultId;
            latestVersion = version;
          }
        });

        let checks = [];
        if (latestVersion.checks?.length > 0) {
          checks = latestVersion.checks.map((check, index) => {
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

        const workflowId = currentVersion.workflow_id;
        const workflowName = currentVersion.workflow_name;

        return {
          name: {
            name: artifactName,
            url: `${getPathPrefix()}/workflow/${workflowId}/result/${latestDagResultId}/artifact/${artifactId}`,
            status: latestVersion.status,
          },
          created_at: new Date(latestVersion.timestamp * 1000).toLocaleString(),
          workflow: {
            name: workflowName,
            url: `${getPathPrefix()}/workflow/${workflowId}`,
            // TODO: Get latest workflow version and show status.
            status: ExecutionStatus.Succeeded,
          },
          // TODO: Get python data type from API route
          type: 'pandas.DataFrame',
          // TODO: Get API route to return metrics in addition to checks array.
          metrics: metricsShort,
          checks: checks,
        };
      }
    );
  }

  // TODO: Change this type to something more generic.
  // Also make this change in WorkflowsTable, I think we can just use Data here if we add JSX.element to Data's union type.
  const mockData: WorkflowTableData = {
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

  return (
    <Layout
      breadcrumbs={[BreadcrumbLink.HOME, BreadcrumbLink.DATA]}
      user={user}
    >
      <Box>
        <WorkflowTable
          data={mockData}
          searchEnabled={true}
          onGetColumnValue={onGetColumnValue}
        />
      </Box>
    </Layout>
  );
};

export default DataPage;
