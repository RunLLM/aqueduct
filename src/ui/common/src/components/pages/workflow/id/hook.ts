import { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useParams } from 'react-router-dom';

import { BreadcrumbLink } from '../../../../components/layouts/NavBar';
import {
  useNodesGetQuery,
  useNodesResultsGetQuery,
} from '../../../../handlers/AqueductApi';
import { handleGetWorkflowDag } from '../../../../handlers/getWorkflowDag';
import { handleGetWorkflowDagResult } from '../../../../handlers/getWorkflowDagResult';
import { NodeResultsMap, NodesMap } from '../../../../handlers/responses/node';
import { WorkflowDagResultWithLoadingStatus } from '../../../../reducers/workflowDagResults';
import { WorkflowDagWithLoadingStatus } from '../../../../reducers/workflowDags';
import { AppDispatch, RootState } from '../../../../stores/store';
import { getPathPrefix } from '../../../../utils/getPathPrefix';
import { isInitial } from '../../../../utils/shared';

export type useWorkflowOutputs = {
  breadcrumbs: BreadcrumbLink[];
  workflowId: string;
  workflowDagId: string;
  workflowDagResultId: string;
  workflowDagWithLoadingStatus: WorkflowDagWithLoadingStatus;
  workflowDagResultWithLoadingStatus: WorkflowDagResultWithLoadingStatus;
};

export default function useWorkflow(
  apiKey: string,
  workflowIdProp: string,
  workflowDagIdProp: string,
  workflowDagResultIdProp: string,
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
