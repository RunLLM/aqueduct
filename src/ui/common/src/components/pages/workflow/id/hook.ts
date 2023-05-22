import { useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';

import { BreadcrumbLink } from '../../../../components/layouts/NavBar';
import {
  useDagResultsGetQuery,
  useDagsGetQuery,
  useNodesGetQuery,
  useNodesResultsGetQuery,
  useWorkflowGetQuery,
} from '../../../../handlers/AqueductApi';
import { NodeResultsMap, NodesMap } from '../../../../handlers/responses/node';
import { DagResultResponse } from '../../../../handlers/responses/workflow';
import { getPathPrefix } from '../../../../utils/getPathPrefix';

export type useWorkflowIdsOutputs = {
  workflowId: string;
  dagId?: string;
  dagResultId?: string;
};

export function useSortedDagResults(
  apiKey: string,
  workflowId: string
): DagResultResponse[] {
  const { data: dagResults } = useDagResultsGetQuery(
    { apiKey, workflowId: workflowId },
    { skip: !workflowId }
  );

  if (!dagResults) {
    return [];
  }

  return [...dagResults]
    .sort((x, y) =>
      new Date(x.exec_state.timestamps?.pending_at) <
      new Date(y.exec_state.timestamps?.pending_at)
        ? -1
        : 1
    )
    .reverse();
}

// useWorkflowIds ensures we use the URL parameter as ground-truth for fetching
// workflow, dag, and result IDs. It includes additional hooks to ensure
// redux states are in-sync.
// This hook should be used for all pages that need to access a single DAG (or DAG result)
// data.
export function useWorkflowIds(apiKey: string): useWorkflowIdsOutputs {
  const navigate = useNavigate();
  const {
    workflowId: wfIdParam,
    dagId: dagIdParam,
    dagResultId: dagResultIdParam,
  } = useParams();

  const dagResults = useSortedDagResults(apiKey, wfIdParam);
  const { isSuccess: dagResultsSuccess } = useDagResultsGetQuery(
    { apiKey, workflowId: wfIdParam },
    { skip: !wfIdParam }
  );

  const { data: dags } = useDagsGetQuery(
    { apiKey, workflowId: wfIdParam },
    // skip if:
    //  1) wfId is not available
    //  2) already have results
    //  3) results not successfully loaded
    {
      skip:
        !wfIdParam ||
        (!!dagResults && dagResults.length > 0) ||
        !dagResultsSuccess,
    }
  );

  // Select the first availale dag result if ID is not provided.
  const dagResult = dagResultIdParam
    ? (dagResults ?? []).filter((r) => r.id === dagResultIdParam)[0]
    : (dagResults ?? [])[0];

  const dag = dagIdParam
    ? (dags ?? []).filter((d) => d.id === dagIdParam)[0]
    : (dags ?? [])[0];

  useEffect(() => {
    if (dagResult !== undefined && !dagResultIdParam) {
      navigate(
        `/workflow/${encodeURI(wfIdParam)}/result/${encodeURI(dagResult.id)}`,
        { replace: true }
      );
      return;
    }

    if (dag !== undefined && !dagIdParam) {
      navigate(`/workflow/${encodeURI(wfIdParam)}/dag/${encodeURI(dag.id)}`, {
        replace: true,
      });
      return;
    }
  }, [wfIdParam, dagResultIdParam, dagResult, dagIdParam, dag]);

  return {
    workflowId: wfIdParam,
    // we take dagResult as first priority
    // otherwise, the workflow has no result and we fall back
    // to dags.
    // If neither are available, either things are still loading
    // or the provided ID is not available.
    dagId: dagResult?.dag_id ?? dag?.id,
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

export function useWorkflowNodes(
  apiKey: string,
  workflowId: string,
  dagId: string | undefined
): NodesMap {
  const { data: nodes } = useNodesGetQuery(
    { apiKey, workflowId, dagId },
    { skip: !workflowId || !dagId }
  );
  return {
    operators: Object.fromEntries(
      (nodes?.operators ?? []).map((op) => [op.id, op])
    ),
    artifacts: Object.fromEntries(
      (nodes?.artifacts ?? []).map((artf) => [artf.id, artf])
    ),
  };
}

export function useWorkflowNodesResults(
  apiKey: string,
  workflowId: string,
  dagResultId: string | undefined
): NodeResultsMap {
  const { data: nodeResults } = useNodesResultsGetQuery(
    { apiKey, workflowId, dagResultId },
    { skip: !workflowId || !dagResultId }
  );
  return {
    operators: Object.fromEntries(
      (nodeResults?.operators ?? []).map((op) => [op.operator_id, op])
    ),
    artifacts: Object.fromEntries(
      (nodeResults?.artifacts ?? []).map((artf) => [artf.artifact_id, artf])
    ),
  };
}
