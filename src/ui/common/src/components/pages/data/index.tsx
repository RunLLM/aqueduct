import { Typography } from '@mui/material';
import Box from '@mui/material/Box';
import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import WorkflowTable, { WorkflowTableData } from '../../../components/tables/WorkflowTable';

import { BreadcrumbLink } from '../../../components/layouts/NavBar';
import { getDataArtifactPreview } from '../../../reducers/dataPreview';
import { handleLoadIntegrations } from '../../../reducers/integrations';
import { AppDispatch, RootState } from '../../../stores/store';
import UserProfile from '../../../utils/auth';
import { DataPreviewInfo } from '../../../utils/data';
import { DataCard } from '../../integrations/cards/card';
import { Card, CardPadding } from '../../layouts/card';
import DefaultLayout from '../../layouts/default';
import { filteredList, SearchBar } from '../../Search';
import { LayoutProps } from '../types';
import CheckItem, { CheckPreview } from '../workflows/components/CheckItem';
import ExecutionStatusLink from '../workflows/components/ExecutionStatusLink';
import MetricItem, { MetricPreview } from '../workflows/components/MetricItem';
import ExecutionStatus from '../../../utils/shared';
import { CheckLevel } from '../../../utils/operators';
import getPathPrefix from '../../../utils/getPathPrefix';

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

  const [filterText, setFilterText] = useState<string>('');

  const displayFilteredCards = (filteredDataCards, idx) => {
    return (
      <Card key={idx} my={2}>
        <DataCard dataPreviewInfo={filteredDataCards} />
      </Card>
    );
  };

  const noItemsMessage = (
    <Typography variant="h5">There are no data artifacts yet.</Typography>
  );

  const dataCards = filteredList(
    filterText,
    Object.values(dataCardsInfo.data.latest_versions),
    (dataCardInfo: DataPreviewInfo) => dataCardInfo.artifact_name,
    displayFilteredCards,
    noItemsMessage
  );

  useEffect(() => {
    document.title = 'Data | Aqueduct';
  }, []);

  const getOptionLabel = (option: DataPreviewInfo) => {
    // When option string is invalid, none of 'options' will be selected
    // and the component will try to directly render the input string.
    // This check prevents applying `dataCardName` to the string.
    if (typeof option === 'string') {
      return option;
    }
    return option.artifact_name;
  };

  // Start of new data table work
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
    tableData = Object.keys(dataCardsInfo.data.latest_versions).map((version) => {
      let currentVersion = dataCardsInfo.data.latest_versions[version.toString()];

      const artifactId = currentVersion.artifact_id;
      const artifactName = currentVersion.artifact_name;

      // TODO: handle the versions array
      // The key ofeach object inside the versions map is a dagResultId
      const dataPreviewInfoVersions = Object.entries(currentVersion.versions);
      let [latestDagResultId, latestVersion] = dataPreviewInfoVersions.length > 0 ? dataPreviewInfoVersions[0] : null;

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
          // TODO: Figure out where to get the checkLevel from.
          console.log('check inside loop: ', check);
          const level = check.metadata.failure_type ? CheckLevel.Warning : CheckLevel.Error;
          const value = check.metadata.status === 'succeeded' && !check.metadata.error;
          return {
            checkId: index,
            name: check.name,
            status: check.status,
            level,
            value: value ? 'True' : 'False',
            timestamp: check.metadata.timestamps.finished_at,
          };
        })
      }

      // console.log('CHECKS: ', checks);

      const workflowDagResultId = currentVersion.workflow_dag_result_id;
      console.log('workflowDagResultId: ', workflowDagResultId);

      const workflowId = currentVersion.workflow_id;
      console.log('workflowId: ', workflowId);

      const workflowName = currentVersion.workflow_name;
      console.log('workflowName: ', workflowName);

      // TODO: Normalize and return the data so that we have an array of rows to render in the table.
      return {
        name: {
          name: artifactName,
          url: `${getPathPrefix()}/workflow/${workflowId}/result/${latestDagResultId}/artifact/${artifactId}`,
          status: ExecutionStatus.Running,
        },
        created_at: new Date(latestVersion.timestamp * 1000).toLocaleString(),
        workflow: {
          name: workflowName,
          url: `${getPathPrefix()}/workflow/${workflowId}`,
          status: ExecutionStatus.Succeeded,
        },
        type: 'pandas.DataFrame',
        // TODO: handle empty array for metrics and checks
        //metrics: [],
        //checks: [],
        meta: [],
        metrics: metricsShort,
        checks: checks,
      }
    });
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
    // data: [
    //   {
    //     name: {
    //       name: 'churn_model',
    //       url: '/data',
    //       status: ExecutionStatus.Succeeded,
    //     },
    //     created_at: '11/1/2022 2:00PM',
    //     workflow: {
    //       name: 'train_churn_model',
    //       url: '/workflows',
    //       status: ExecutionStatus.Running,
    //     },
    //     type: 'sklearn.linear, Linear Regression',
    //     metrics: metricsShort,
    //     checks: checkPreviews,
    //   },
    //   {
    //     name: {
    //       name: 'predict_churn_dataset',
    //       url: '/workflows',
    //       status: ExecutionStatus.Running,
    //     },
    //     created_at: '11/1/2022 2:00PM',
    //     workflow: {
    //       name: 'monthly_churn_prediction',
    //       url: '/workflows',
    //       status: ExecutionStatus.Succeeded,
    //     },
    //     type: 'pandas.DataFrame',
    //     metrics: metricsShort,
    //     checks: checkPreviews,
    //   },
    //   {
    //     name: {
    //       name: 'label_classifier',
    //       url: '/data',
    //       status: ExecutionStatus.Pending,
    //     },
    //     created_at: '11/1/2022 2:00PM',
    //     workflow: {
    //       name: 'label_classifier_workflow',
    //       url: '/workflows',
    //       status: ExecutionStatus.Registered,
    //     },
    //     type: 'parquet',
    //     metrics: metricsShort,
    //     checks: checkPreviews,
    //   },
    // ],
    // TODO: Remove this meta array, it's no longer needed.
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

  // TODO: Figure out how to show the integration in which the data is stored.
  // See Integration card for inspiration there.
  // Like in the workflow table, we can include a small icon for the integration and show the name of the integration
  // next to the icon.

  // return (
  //   <Layout
  //     breadcrumbs={[BreadcrumbLink.HOME, BreadcrumbLink.DATA]}
  //     user={user}
  //   >
  //     <Box>
  //       <WorkflowTable
  //         data={mockData}
  //         searchEnabled={true}
  //         onGetColumnValue={onGetColumnValue}
  //       />
  //     </Box>
  //   </Layout>
  // );


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
      <Box>
        <Box paddingLeft={CardPadding}>
          {/* Aligns search bar to card text */}
          <SearchBar
            options={Object.values(dataCardsInfo.data.latest_versions)}
            getOptionLabel={getOptionLabel}
            setSearchTerm={setFilterText}
          />
        </Box>
        {dataCards}
      </Box>
    </Layout>
  );
};

export default DataPage;
