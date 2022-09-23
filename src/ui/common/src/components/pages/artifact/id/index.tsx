import { CircularProgress } from '@mui/material';
import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React, { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useParams } from 'react-router-dom';
import { handleGetArtifactResultContent } from '../../../../handlers/getArtifactResultContent';

import { handleGetWorkflowDagResult } from '../../../../handlers/getWorkflowDagResult';
import { getMetricsAndChecksOnArtifact } from '../../../../handlers/responses/dag';
import { AppDispatch, RootState } from '../../../../stores/store';
import { ArtifactType } from '../../../../utils/artifacts';
import UserProfile from '../../../../utils/auth';
import { DataSchema } from '../../../../utils/data';
import { exportCsv } from '../../../../utils/preview';
import { isFailed, isInitial, isLoading } from '../../../../utils/shared';
import DefaultLayout from '../../../layouts/default';
import { Button } from '../../../primitives/Button.styles';
import OperatorExecStateTable, {
  OperatorExecStateTableType,
} from '../../../tables/OperatorExecStateTable';
import PaginatedTable from '../../../tables/PaginatedTable';
import DetailsPageHeader from '../../components/DetailsPageHeader';
import { LayoutProps } from '../../types';

type ArtifactDetailsPageProps = {
  user: UserProfile;
  Layout?: React.FC<LayoutProps>;
};

const ArtifactDetailsPage: React.FC<ArtifactDetailsPageProps> = ({
  user,
  Layout = DefaultLayout,
}) => {
  const dispatch: AppDispatch = useDispatch();
  const { workflowId, workflowDagResultId, artifactId } = useParams();
  const workflowDagResultWithLoadingStatus = useSelector(
    (state: RootState) =>
      state.workflowDagResultsReducer.results[workflowDagResultId]
  );
  const artifactContents = useSelector(
    (state: RootState) => state.artifactResultContentsReducer.contents
  );
  const artifact = (workflowDagResultWithLoadingStatus?.result?.artifacts ??
    {})[artifactId];
  const artifactResultId = artifact?.result?.id;
  const contentsWithLoadingStatus = artifactResultId
    ? artifactContents[artifactResultId]
    : undefined;
  const { metrics, checks } = !!workflowDagResultWithLoadingStatus
    ? getMetricsAndChecksOnArtifact(
        workflowDagResultWithLoadingStatus?.result,
        artifactId
      )
    : { metrics: [], checks: [] };

  useEffect(() => {
    document.title = 'Artifact Details | Aqueduct';

    // Load workflow dag result if it's not cached
    if (
      !workflowDagResultWithLoadingStatus ||
      isInitial(workflowDagResultWithLoadingStatus.status)
    ) {
      dispatch(
        handleGetWorkflowDagResult({
          apiKey: user.apiKey,
          workflowId,
          workflowDagResultId,
        })
      );
    }
  }, []);

  useEffect(() => {
    if (!!artifact) {
      document.title = `${artifact.name} | Aqueduct`;

      if (
        !!artifact.result &&
        artifact.result.content_serialized === undefined &&
        !contentsWithLoadingStatus
      ) {
        dispatch(
          handleGetArtifactResultContent({
            apiKey: user.apiKey,
            artifactId,
            artifactResultId,
            workflowDagResultId,
          })
        );
      }
    }
  }, [artifact]);

  if (
    !workflowDagResultWithLoadingStatus ||
    !isInitial(workflowDagResultWithLoadingStatus.status) ||
    !isLoading(workflowDagResultWithLoadingStatus.status)
  ) {
    return (
      <Layout user={user}>
        <CircularProgress />
      </Layout>
    );
  }

  if (isFailed(workflowDagResultWithLoadingStatus.status)) {
    return (
      <Layout user={user}>
        <Alert title="Failed to load workflow">
          {workflowDagResultWithLoadingStatus.status.err}
        </Alert>
      </Layout>
    );
  }

  if (!artifact) {
    return (
      <Layout user={user}>
        <Alert title="Failed to load artifact">
          Artifact {artifactId} doesn't seem to exist on this workflow.
        </Alert>
      </Layout>
    );
  }

  const metricsChecksSchema: DataSchema = {
    fields: [
      { name: 'Title', type: 'varchar' },
      { name: 'Value', type: 'varchar' },
    ],
    pandas_version: '0.0.1', // TODO: Figure out what to set this value to.
  };

  const metricTableEntries = {
    schema: metricsChecksSchema,
    data: metrics
      .map((metricArtf) => {
        let name = metricArtf.name;
        if (name.endsWith('artifact') || name.endsWith('Aritfact')) {
          name = name.slice(0, 'artifact'.length);
        }
        return {
          name: name,
          value: metricArtf.result?.content_serialized,
        };
      })
      .filter((x) => !!x.value),
  };

  const metricsSection = (
    <Box width="100%">
      <Typography variant="h5" component="div" marginBottom="8px">
        Metrics
      </Typography>
      {metricTableEntries.data.length > 0 ? (
        <OperatorExecStateTable
          schema={metricTableEntries.schema}
          rows={metricTableEntries}
          tableType={OperatorExecStateTableType.Metric}
        />
      ) : (
        <Typography variant="body2">
          This artifact has no associated downstream Metrics.
        </Typography>
      )}
    </Box>
  );

  const checkTableEntries = {
    schema: metricsChecksSchema,
    data: checks
      .map((checkArtf) => {
        let name = checkArtf.name;
        if (name.endsWith('artifact') || name.endsWith('Aritfact')) {
          name = name.slice(0, 'artifact'.length);
        }
        return {
          name: name,
          value: checkArtf.result?.content_serialized,
        };
      })
      .filter((x) => !!x.value),
  };

  const checksSection = (
    <Box width="100%">
      <Typography variant="h5" component="div" marginBottom="8px">
        Checks
      </Typography>
      {checkTableEntries.data.length > 0 ? (
        <OperatorExecStateTable
          schema={checkTableEntries.schema}
          rows={checkTableEntries}
          tableType={OperatorExecStateTableType.Check}
        />
      ) : (
        <Typography variant="body2">
          This artifact has no associated downstream Checks.
        </Typography>
      )}
    </Box>
  );

  let csvExporter = null;
  let contentSection = null;
  if (!artifact.result) {
    contentSection = (
      <Typography variant="h5" component="div" marginBottom="8px">
        No result to show for this artifact.
      </Typography>
    );
  } else if (artifact.result.content_serialized !== undefined) {
    contentSection = (
      <Typography variant="body1" component="div" marginBottom="8px">
        <code>{artifact.result.content_serialized}</code>
      </Typography>
    );
  } else if (!!contentsWithLoadingStatus) {
    if (
      isInitial(contentsWithLoadingStatus.status) ||
      isLoading(contentsWithLoadingStatus.status)
    ) {
      contentSection = <CircularProgress />;
    } else if (isFailed(contentsWithLoadingStatus.status)) {
      contentSection = (
        <Alert title="Failed to load artifact contents.">
          {contentsWithLoadingStatus.status.err}
        </Alert>
      );
    } else if (!contentsWithLoadingStatus.data) {
      contentSection = (
        <Typography variant="h5" component="div" marginBottom="8px">
          No result to show for this artifact.
        </Typography>
      );
    } else {
      if (
        artifact.type === ArtifactType.Bytes ||
        artifact.type === ArtifactType.Picklable
      ) {
        contentSection = (
          <Button
            variant="contained"
            sx={{ maxHeight: '32px' }}
            onClick={() => {
              const content = contentsWithLoadingStatus.data;
              const blob = new Blob([content], { type: 'text' });
              const url = window.URL.createObjectURL(blob);
              const a = document.createElement('a');
              a.href = url;
              a.download = artifact.name;
              a.click();

              return true;
            }}
          >
            Download
          </Button>
        );
      } else if (artifact.type === ArtifactType.Table) {
        try {
          const data = JSON.parse(contentsWithLoadingStatus.data);
          csvExporter = (
            <Button
              variant="contained"
              sx={{ maxHeight: '32px' }}
              onClick={() => {
                exportCsv(
                  data,
                  artifact.name ? artifact.name.replaceAll(' ', '_') : 'data'
                );
              }}
            >
              Export
            </Button>
          );
          contentSection = <PaginatedTable data={data} />;
        } catch (err) {
          contentSection = (
            <Alert title="Cannot parse table data.">
              {err}
              {contentsWithLoadingStatus.data}
            </Alert>
          );
        }
      } else {
        // TODO: handle images here
        contentSection = (
          <Typography variant="body1" component="div" marginBottom="8px">
            <code>{contentsWithLoadingStatus.data}</code>
          </Typography>
        );
      }
    }
  } else {
    contentSection = <CircularProgress />;
  }

  return (
    <Layout user={user}>
      <Box width={'800px'}>
        <Box width="100%">
          <Box width="100%" display="flex" alignItems="center">
            <DetailsPageHeader name={artifact.name} />
            {csvExporter}
          </Box>
          <Box width="100%" marginTop="12px">
            <Typography variant="h5" component="div" marginBottom="8px">
              Preview
            </Typography>
            {contentSection}
          </Box>
          <Box display="flex" width="100%" paddingTop="40px">
            {metricsSection}
            <Box width="96px" />
            {checksSection}
          </Box>
        </Box>
      </Box>
    </Layout>
  );
};

export default ArtifactDetailsPage;
