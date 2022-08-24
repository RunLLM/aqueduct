import { faCircleCheck } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Button, CircularProgress } from '@mui/material';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React, { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useNavigate, useParams } from 'react-router-dom';
import { Data, DataSchema } from 'src/utils/data';

import {
  ArtifactResult,
  handleGetArtifactResults,
} from '../../../../reducers/workflow';
import { AppDispatch, RootState } from '../../../../stores/store';
import UserProfile from '../../../../utils/auth';
import { exportCsv } from '../../../../utils/preview';
import DefaultLayout from '../../../layouts/default';
import KeyValueTable, {
  KeyValueTableType,
} from '../../../tables/KeyValueTable';
import StickyHeaderTable from '../../../tables/StickyHeaderTable';
import { LayoutProps } from '../../types';

const kvSchema: DataSchema = {
  fields: [
    { name: 'Title', type: 'varchar' },
    { name: 'Value', type: 'varchar' },
  ],
  pandas_version: '0.0.1', // TODO: Figure out what to set this value to.
};

const mockMetrics: Data = {
  schema: kvSchema,
  data: [
    ['avg_churn', '0.04'],
    ['avg_workflows', '455'],
    ['avg_users', '1.2'],
    ['avg_users', '5'],
  ],
};

const mockChecks: Data = {
  schema: kvSchema,
  data: [
    ['reasonable_churn', 'True'],
    ['wf_count_small', 'False'],
    ['bounds_check', 'True'],
    ['avg_users_check', 'False'],
    ['warning_check', 'Warning'],
    ['none_check', 'None'],
  ],
};

type ArtifactDetailsHeaderProps = {
  artifactName: string;
  createdAt?: string;
  sourceLocation?: string;
};

const ArtifactDetailsHeader: React.FC<ArtifactDetailsHeaderProps> = ({
  artifactName,
  // TODO: add these back once we have support for getting createdAt and sourceLocation.
  //createdAt,
  //sourceLocation,
}) => {
  return (
    <Box width="100%" display="flex" alignItems="center">
      <FontAwesomeIcon
        height="24px"
        width="24px"
        style={{ marginRight: '8px' }}
        icon={faCircleCheck}
        color={'green'}
      />
      <Typography variant="h4" component="div">
        {artifactName}
      </Typography>
      {/* <Typography marginTop="4px" variant="caption" component="div">
                Created: {createdAt}
            </Typography>
            <Typography variant="caption" component="div">
                Source: <Link>{sourceLocation}</Link>
            </Typography> */}
    </Box>
  );
};

type ArtifactDetailsPageProps = {
  user: UserProfile;
  Layout?: React.FC<LayoutProps>;
};

const ArtifactDetailsPage: React.FC<ArtifactDetailsPageProps> = ({
  user,
  Layout = DefaultLayout,
}) => {
  const workflow = useSelector((state: RootState) => state.workflowReducer);
  const navigate = useNavigate();
  const dispatch: AppDispatch = useDispatch();
  const { workflowId, workflowDagResultId, artifactId } = useParams();
  const artifactResult: ArtifactResult | null = useSelector(
    (state: RootState) => {
      // First, check if there are any keys in the artifactResults object.
      const artifactResults = state.workflowReducer.artifactResults;
      if (Object.keys(artifactResults).length < 1) {
        return null;
      }

      return artifactResults[artifactId];
    }
  );

  useEffect(() => {
    document.title = 'Artifact | Aqueduct';

    // Check and see if we are loading the artifact result
    if (!artifactResult) {
      dispatch(
        handleGetArtifactResults({
          apiKey: user.apiKey,
          workflowDagResultId,
          artifactId,
        })
      );
    }
  }, []);

  if (!artifactResult || !artifactResult.result) {
    return (
      <Layout user={user}>
        <CircularProgress />
      </Layout>
    );
  }

  const parsedData = JSON.parse(artifactResult.result.data);
  const artifactName: string = artifactResult.result.name;

  return (
    <Layout user={user}>
      <Box width={'800px'}>
        <Box width="100%">
          <Box width="100%" display="flex" alignItems="center">
            <ArtifactDetailsHeader artifactName={artifactName} />
            <Button
              variant="contained"
              sx={{ maxHeight: '32px' }}
              onClick={() => {
                exportCsv(
                  parsedData,
                  artifactName ? artifactName.replaceAll(' ', '_') : 'data'
                );
              }}
            >
              Export
            </Button>
          </Box>
          <Box width="100%" marginTop="12px">
            <Typography variant="h5" component="div" marginBottom="8px">
              Preview
            </Typography>
            <StickyHeaderTable data={parsedData} />
          </Box>
          <Box display="flex" width="100%" paddingTop="40px">
            <Box width="100%">
              <Typography variant="h5" component="div" marginBottom="8px">
                Metrics
              </Typography>
              {mockMetrics.data.length > 0 ? (
                <KeyValueTable
                  schema={kvSchema}
                  rows={mockMetrics}
                  tableType={KeyValueTableType.Metric}
                />
              ) : (
                <Typography variant="body2">
                  This artifact has no associated downstream Metrics.
                </Typography>
              )}
            </Box>
            <Box width="96px" />
            <Box width="100%">
              <Typography variant="h5" component="div" marginBottom="8px">
                Checks
              </Typography>
              {mockChecks.data.length > 0 ? (
                <KeyValueTable
                  schema={kvSchema}
                  rows={mockChecks}
                  tableType={KeyValueTableType.Check}
                />
              ) : (
                <Typography variant="body2">
                  This artifact has no associated downstream Checks.
                </Typography>
              )}
            </Box>
          </Box>
        </Box>
      </Box>
    </Layout>
  );
};

export default ArtifactDetailsPage;
