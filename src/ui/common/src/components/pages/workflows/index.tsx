import { Link, Typography } from '@mui/material';
import Box from '@mui/material/Box';
import Divider from '@mui/material/Divider';
import React, { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { handleFetchAllWorkflowSummaries } from '../../../reducers/listWorkflowSummaries';
import { AppDispatch, RootState } from '../../../stores/store';
import UserProfile from '../../../utils/auth';
import { LoadingStatusEnum } from '../../../utils/shared';
import DefaultLayout from '../../layouts/default';
import WorkflowCard from '../../workflows/workflowCard';

type Props = {
  user: UserProfile;
};

const WorkflowsPage: React.FC<Props> = ({ user }) => {
  const dispatch: AppDispatch = useDispatch();

  useEffect(() => {
    dispatch(handleFetchAllWorkflowSummaries({ apiKey: user.apiKey }));
  }, []);

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

  const heading = (
    <Box mb={2}>
      <Typography variant="h2" gutterBottom component="div">
        Workflows
      </Typography>
    </Box>
  );

  const workflowList =
    allWorkflows.workflows.length > 0 ? (
      <Box sx={{ maxWidth: '1000px', width: '90%' }}>
        {allWorkflows.workflows.map((workflow, idx) => {
          return (
            <React.Fragment key={idx}>
              <Box my={2}>
                <WorkflowCard workflow={workflow} />
              </Box>
              {idx < allWorkflows.workflows.length - 1 && <Divider />}
            </React.Fragment>
          );
        })}
      </Box>
    ) : (
      <Typography variant="h5">
        There are no workflows created yet. Create one right now with our{' '}
        <Link href="https://github.com/aqueducthq/aqueduct/blob/main/sdk">
          Python SDK
        </Link>
        !
      </Typography>
    );

  return (
    <DefaultLayout user={user}>
      <></>
      <Box p={2}>
        {heading}
        {workflowList}
      </Box>
    </DefaultLayout>
  );
};

export default WorkflowsPage;
