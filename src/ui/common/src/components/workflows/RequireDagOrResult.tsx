import { Alert, AlertTitle, CircularProgress } from '@mui/material';
import React from 'react';

import { WorkflowDagResultWithLoadingStatus } from '../../reducers/workflowDagResults';
import { WorkflowDagWithLoadingStatus } from '../../reducers/workflowDags';
import {
  isFailed,
  isInitial,
  isLoading,
  isSucceeded,
} from '../../utils/shared';

type Props = {
  dagWithLoadingStatus?: WorkflowDagWithLoadingStatus;
  dagResultWithLoadingStatus?: WorkflowDagResultWithLoadingStatus;
  children: React.ReactElement | React.ReactElement[];
};

const RequireDagOrResult: React.FC<Props> = ({
  dagWithLoadingStatus,
  dagResultWithLoadingStatus,
  children,
}) => {
  // This workflow doesn't exist.
  if (dagResultWithLoadingStatus) {
    if (
      isInitial(dagResultWithLoadingStatus.status) ||
      isLoading(dagResultWithLoadingStatus.status)
    ) {
      return <CircularProgress />;
    }

    if (isFailed(dagResultWithLoadingStatus.status)) {
      <Alert severity="error">
        <AlertTitle>Failed to load dag reult.</AlertTitle>
        {dagResultWithLoadingStatus.status.err}
      </Alert>;
    }

    if (isSucceeded(dagResultWithLoadingStatus.status)) {
      return <>{children}</>;
    }
  }

  if (dagWithLoadingStatus) {
    if (
      isInitial(dagWithLoadingStatus.status) ||
      isLoading(dagWithLoadingStatus.status)
    ) {
      return <CircularProgress />;
    }

    if (isFailed(dagWithLoadingStatus.status)) {
      <Alert severity="error">
        <AlertTitle>Failed to load dag.</AlertTitle>
        {dagWithLoadingStatus.status.err}
      </Alert>;
    }

    if (isSucceeded(dagWithLoadingStatus.status)) {
      return <>{children}</>;
    }
  }

  return <CircularProgress />;
};

export default RequireDagOrResult;
