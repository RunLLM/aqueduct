import { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useNavigate, useParams } from 'react-router-dom';

import { BreadcrumbLink } from '../../../../components/layouts/NavBar';
import { useDagResultsGetQuery } from '../../../../handlers/AqueductApi';
import { handleGetWorkflowDag } from '../../../../handlers/getWorkflowDag';
import { handleGetWorkflowDagResult } from '../../../../handlers/getWorkflowDagResult';
import { initializeDagOrResultPageIfNotExists } from '../../../../reducers/pages/Workflow';
import { WorkflowDagResultWithLoadingStatus } from '../../../../reducers/workflowDagResults';
import { WorkflowDagWithLoadingStatus } from '../../../../reducers/workflowDags';
import { AppDispatch, RootState } from '../../../../stores/store';
import { getPathPrefix } from '../../../../utils/getPathPrefix';
import { isInitial } from '../../../../utils/shared';

export type useWorkflowIdsOutputs = {
  workflowId: string;
  dagId?: string;
  dagResultId?: string;
};

export type useWorkflowOutputs = {
  breadcrumbs: BreadcrumbLink[];
  workflowId: string;
  workflowDagId: string;
  workflowDagResultId: string;
  workflowDagWithLoadingStatus: WorkflowDagWithLoadingStatus;
  workflowDagResultWithLoadingStatus: WorkflowDagResultWithLoadingStatus;
};

// useWorkflowIds ensures we use the URL parameter as ground-truth for fetching
// workflow, dag, and result IDs. It includes additional hooks to ensure
// redux states are in-sync.
// This hook should be used for all pages that need to access a single DAG (or DAG result)
// data.
export function useWorkflowIds(apiKey: string): useWorkflowIdsOutputs {
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const {
    id: wfIdParam,
    dagId: dagIdParam,
    dagResultId: dagResultIdParam,
  } = useParams();

  const { data: dagResults } = useDagResultsGetQuery(
    { apiKey, workflowId: wfIdParam },
    { skip: !wfIdParam }
  );

  // Select the first availale dag result if ID is not provided.
  const dagResult = dagResultIdParam
    ? (dagResults ?? []).filter((r) => r.id === dagResultIdParam)[0]
    : dagResults[0];

  useEffect(() => {
    if (dagResult !== undefined) {
      dispatch(
        initializeDagOrResultPageIfNotExists({
          workflowId: wfIdParam,
          dagId: dagResult.dag_id,
          dagResultId: dagResult.id,
        })
      );

      if (!dagResultIdParam) {
        navigate(`result/${encodeURI(dagResult.id)}`, { replace: true });
      }
    }
  }, [wfIdParam, dagResultIdParam, dagResult]);

  return {
    workflowId: wfIdParam,
    dagId: dagResult?.dag_id ?? dagIdParam,
    dagResultId: dagResult?.id,
  };
}

export default function useWorkflow(
  apiKey: string,
  workflowIdProp: string | undefined,
  workflowDagIdProp: string | undefined,
  workflowDagResultIdProp: string | undefined,
  title = 'Workflow'
): useWorkflowOutputs {
  const dispatch: AppDispatch = useDispatch();
  let { workflowId, workflowDagId, workflowDagResultId } = useParams();

  if (workflowIdProp) {
    workflowId = workflowIdProp;
  }

  if (workflowDagIdProp) {
    workflowDagId = workflowDagIdProp;
  }

  if (workflowDagResultIdProp) {
    workflowDagResultId = workflowDagResultIdProp;
  }

  const workflowDagResultWithLoadingStatus = useSelector(
    (state: RootState) =>
      state.workflowDagResultsReducer.results[workflowDagResultId]
  );

  const workflowDagWithLoadingStatus = useSelector(
    (state: RootState) => state.workflowDagsReducer.results[workflowDagId]
  );

  const pathPrefix = getPathPrefix();
  const workflowLink = `${pathPrefix}/workflow/${workflowId}?workflowDagResultId=${workflowDagResultId}`;
  const breadcrumbs = [
    BreadcrumbLink.HOME,
    BreadcrumbLink.WORKFLOWS,
    new BreadcrumbLink(
      workflowLink,
      workflowDagResultWithLoadingStatus?.result?.name ?? title
    ),
  ];

  useEffect(() => {
    if (
      // Load workflow dag result if it's not cached
      (!workflowDagResultWithLoadingStatus ||
        isInitial(workflowDagResultWithLoadingStatus.status)) &&
      workflowDagResultId
    ) {
      dispatch(
        handleGetWorkflowDagResult({
          apiKey: apiKey,
          workflowId,
          workflowDagResultId,
        })
      );
    }

    if (
      (!workflowDagWithLoadingStatus ||
        isInitial(workflowDagWithLoadingStatus.status)) &&
      !workflowDagResultId &&
      workflowDagId
    ) {
      dispatch(
        handleGetWorkflowDag({ apiKey: apiKey, workflowId, workflowDagId })
      );
    }
  }, [
    dispatch,
    apiKey,
    workflowDagResultId,
    workflowDagId,
    workflowDagWithLoadingStatus,
    workflowDagResultWithLoadingStatus,
    workflowId,
  ]);

  return {
    breadcrumbs,
    workflowId,
    workflowDagId,
    workflowDagResultId,
    workflowDagWithLoadingStatus,
    workflowDagResultWithLoadingStatus,
  };
}
