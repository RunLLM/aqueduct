import { useEffect } from 'react';
import { useLocation, useParams } from 'react-router-dom';

import { BreadcrumbLink } from '../../../../components/layouts/NavBar';
import { OperatorResultResponse } from '../../../../handlers/responses/operator';
import { WorkflowDagResultWithLoadingStatus } from '../../../../reducers/workflowDagResults';
import { WorkflowDagWithLoadingStatus } from '../../../../reducers/workflowDags';

export type useOperatorOutputs = {
  breadcrumbs: BreadcrumbLink[];
  operatorId: string;
  operator: OperatorResultResponse;
};

export default function useOpeartor(
  id: string,
  workflowBreadcrumbs: BreadcrumbLink[],
  workflowDagWithLoadingStatus: WorkflowDagWithLoadingStatus,
  workflowDagResultWithLoadingStatus: WorkflowDagResultWithLoadingStatus,
  showDocumentTitle: boolean,
  title = 'Operator'
): useOperatorOutputs {
  let { operatorId } = useParams();
  const path = useLocation().pathname;

  if (id) {
    operatorId = id;
  }

  const operator = !!workflowDagResultWithLoadingStatus?.result
    ? (workflowDagResultWithLoadingStatus?.result?.operators ?? {})[operatorId]
    : ((workflowDagWithLoadingStatus?.result?.operators ?? {})[
        operatorId
      ] as OperatorResultResponse);

  const breadcrumbs = [
    ...workflowBreadcrumbs,
    new BreadcrumbLink(path, operator?.name || title),
  ];

  useEffect(() => {
    if (!!operator && showDocumentTitle) {
      document.title = `${operator?.name || title} | Aqueduct`;
    }
  }, [operator, showDocumentTitle, title]);

  return {
    breadcrumbs: breadcrumbs,
    operatorId,
    operator,
  };
}
