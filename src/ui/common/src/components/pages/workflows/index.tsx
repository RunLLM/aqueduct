import { Link, Typography } from '@mui/material';
import Box from '@mui/material/Box';
import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { BreadcrumbLink } from '../../../components/layouts/NavBar';
import { handleFetchAllWorkflowSummaries } from '../../../reducers/listWorkflowSummaries';
import { AppDispatch, RootState } from '../../../stores/store';
import UserProfile from '../../../utils/auth';
import { LoadingStatusEnum } from '../../../utils/shared';
import DefaultLayout from '../../layouts/default';
import { filteredList, SearchBar } from '../../Search';
import WorkflowCard from '../../workflows/workflowCard';
import { LayoutProps } from '../types';

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
  }, []);

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

  const heading = (
    <Box mb={2}>
      <Typography variant="h2" gutterBottom component="div">
        Workflows
      </Typography>
    </Box>
  );

  // TODO: Figure out why we have a _ here, probably don't need it since it's unused.
  const displayFilteredWorkflows = (workflow, _) => {
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

  return (
    <Layout
      breadcrumbs={[BreadcrumbLink.HOME, BreadcrumbLink.WORKFLOWS]}
      user={user}
    >
      <Box p={2}>
        {heading}
        {allWorkflows.workflows.length >= 1 && (
          <SearchBar
            options={allWorkflows.workflows}
            getOptionLabel={(option) => option.name || ''}
            setSearchTerm={setFilterText}
          />
        )}
        {workflowList}
      </Box>
    </Layout>
  );
};

export default WorkflowsPage;
