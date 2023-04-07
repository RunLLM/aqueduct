import { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useLocation, useParams } from 'react-router-dom';

import { BreadcrumbLink } from '../../../../components/layouts/NavBar';
import { handleGetArtifactResultContent } from '../../../../handlers/getArtifactResultContent';
import { handleListArtifactResults } from '../../../../handlers/listArtifactResults';
import { ArtifactResultResponse } from '../../../../handlers/responses/artifactDeprecated';
import { DagResultResponse } from '../../../../handlers/responses/dagDeprecated';
import { ContentWithLoadingStatus } from '../../../../reducers/artifactResultContents';
import { ArtifactResultsWithLoadingStatus } from '../../../../reducers/artifactResults';
import { WorkflowDagResultWithLoadingStatus } from '../../../../reducers/workflowDagResults';
import { WorkflowDagWithLoadingStatus } from '../../../../reducers/workflowDags';
import { AppDispatch, RootState } from '../../../../stores/store';
import { isInitial, isLoading } from '../../../../utils/shared';

export type useArtifactOutputs = {
  breadcrumbs: BreadcrumbLink[];
  artifactId: string;
  artifact: ArtifactResultResponse;
  contentWithLoadingStatus: ContentWithLoadingStatus;
};

export default function useArtifact(
  apiKey: string,
  id: string,
  workflowBreadcrumbs: BreadcrumbLink[],
  workflowDagResultId: string,
  workflowDagWithLoadingStatus: WorkflowDagWithLoadingStatus,
  workflowDagResultWithLoadingStatus: WorkflowDagResultWithLoadingStatus,
  showDocumentTitle: boolean,
  title = 'Artifact Details'
): useArtifactOutputs {
  const dispatch: AppDispatch = useDispatch();
  let { artifactId } = useParams();
  const path = useLocation().pathname;

  if (id) {
    artifactId = id;
  }

  const artifactContents = useSelector(
    (state: RootState) => state.artifactResultContentsReducer.contents
  );

  const dagResult =
    workflowDagResultWithLoadingStatus?.result ??
    (workflowDagWithLoadingStatus?.result as DagResultResponse);
  const artifact = (dagResult?.artifacts ?? {})[artifactId];

  const artifactResultId = artifact?.result?.id;
  const contentWithLoadingStatus = artifactResultId
    ? artifactContents[artifactResultId]
    : undefined;

  const breadcrumbs = [
    ...workflowBreadcrumbs,
    new BreadcrumbLink(path, artifact ? artifact.name : title),
  ];

  useEffect(() => {
    if (!!artifact) {
      if (showDocumentTitle) {
        document.title = `${artifact ? artifact.name : title} | Aqueduct`;
      }

      if (
        !!artifact.result &&
        // intentional '==' to check undefined or null.
        artifact.result.content_serialized == null &&
        !contentWithLoadingStatus
      ) {
        dispatch(
          handleGetArtifactResultContent({
            apiKey: apiKey,
            artifactId,
            artifactResultId,
            workflowDagResultId,
          })
        );
      }
    }
  }, [
    artifact,
    artifactId,
    artifactResultId,
    contentWithLoadingStatus,
    dispatch,
    showDocumentTitle,
    apiKey,
    workflowDagResultId,
    title,
  ]);

  return {
    breadcrumbs,
    artifactId,
    artifact,
    contentWithLoadingStatus,
  };
}

export function useArtifactHistory(
  apiKey: string,
  id: string,
  workflowId: string,
  workflowDagResultWithLoadingStatus: WorkflowDagResultWithLoadingStatus
): ArtifactResultsWithLoadingStatus {
  const dispatch: AppDispatch = useDispatch();
  const artifactHistoryWithLoadingStatus = useSelector((state: RootState) =>
    !!id ? state.artifactResultsReducer.artifacts[id] : undefined
  );

  useEffect(() => {
    // Load artifact history once workflow dag results finished loading
    // and the result is not cached
    if (
      !artifactHistoryWithLoadingStatus &&
      !!id &&
      !!workflowDagResultWithLoadingStatus &&
      !isInitial(workflowDagResultWithLoadingStatus.status) &&
      !isLoading(workflowDagResultWithLoadingStatus.status)
    ) {
      // Queue up the artifacts historical results for loading.
      dispatch(
        handleListArtifactResults({
          apiKey: apiKey,
          workflowId,
          artifactId: id,
        })
      );
    }
  }, [
    workflowDagResultWithLoadingStatus,
    id,
    artifactHistoryWithLoadingStatus,
    dispatch,
    apiKey,
    workflowId,
  ]);

  return artifactHistoryWithLoadingStatus;
}
