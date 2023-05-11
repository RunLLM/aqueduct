import { useEffect } from 'react';
import { useDispatch } from 'react-redux';
import { useNavigate, useParams } from 'react-router-dom';

import { BreadcrumbLink } from '../../../../components/layouts/NavBar';
import {
  useDagResultsGetQuery,
  useWorkflowGetQuery,
} from '../../../../handlers/AqueductApi';
import { initializeDagOrResultPageIfNotExists } from '../../../../reducers/pages/Workflow';
import { getPathPrefix } from '../../../../utils/getPathPrefix';

export type useWorkflowIdsOutputs = {
  workflowId: string;
  dagId?: string;
  dagResultId?: string;
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
    workflowDagId: dagIdParam,
    workflowDagResultId: dagResultIdParam,
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

      if (
        dagIdParam !== dagResult.dag_id ||
        dagResultIdParam !== dagResult.id
      ) {
        navigate(
          `?workflowDagId=${encodeURI(
            dagResult.dag_id
          )}&workflowDagResultId=${encodeURI(dagResult.id)}`,
          { replace: true }
        );
      }
    }
  }, [wfIdParam, dagIdParam, dagResultIdParam, dagResult]);

  return {
    workflowId: wfIdParam,
    dagId: dagResult?.dag_id ?? dagIdParam,
    dagResultId: dagResult?.id,
  };
}

export function useWorkflowBreadcrumbs(
  apiKey: string,
  workflowId: string | undefined,
  dagId: string | undefined,
  dagResultId: string | undefined,
  defaultTitle = 'Workflow'
): BreadcrumbLink[] {
  const { data: workflow } = useWorkflowGetQuery(
    { apiKey, workflowId },
    { skip: !workflowId }
  );

  const pathPrefix = getPathPrefix();
  let workflowLink = `${pathPrefix}/workflow/${workflowId}`;
  if (dagId || dagResultId) {
    workflowLink += '?';
  }

  if (dagId) {
    workflowLink += `workflowDagId=${dagId}`;
  }

  if (dagId && dagResultId) {
    workflowLink += '&';
  }

  if (dagResultId) {
    workflowLink += `workflowDagResultId=${dagResultId}`;
  }

  return [
    BreadcrumbLink.HOME,
    BreadcrumbLink.WORKFLOWS,
    new BreadcrumbLink(workflowLink, workflow?.name ?? defaultTitle),
  ];
}
