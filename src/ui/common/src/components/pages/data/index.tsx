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

  // const dataCards = filteredList(
  //   filterText,
  //   Object.values(dataCardsInfo.data.latest_versions),
  //   (dataCardInfo: DataPreviewInfo) => dataCardInfo.artifact_name,
  //   displayFilteredCards,
  //   noItemsMessage
  // );

  useEffect(() => {
    document.title = 'Data | Aqueduct';
  }, []);

  // const getOptionLabel = (option: DataPreviewInfo) => {
  //   // When option string is invalid, none of 'options' will be selected
  //   // and the component will try to directly render the input string.
  //   // This check prevents applying `dataCardName` to the string.
  //   if (typeof option === 'string') {
  //     return option;
  //   }
  //   return option.artifact_name;
  // };


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


  console.log('dataCardsInfo: ', dataCardsInfo);

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


  // return (
  //   <Layout
  //     breadcrumbs={[BreadcrumbLink.HOME, BreadcrumbLink.DATA]}
  //     user={user}
  //   >
  //     <div />
  //     <Box>
  //       <Box paddingLeft={CardPadding}>
  //         {/* Aligns search bar to card text */}
  //         <SearchBar
  //           options={Object.values(dataCardsInfo.data.latest_versions)}
  //           getOptionLabel={getOptionLabel}
  //           setSearchTerm={setFilterText}
  //         />
  //       </Box>
  //       {dataCards}
  //     </Box>
  //   </Layout>
  // );
};

export default DataPage;
