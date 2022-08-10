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
  handleGetWorkflow,
} from '../../../../reducers/workflow';
import { AppDispatch, RootState } from '../../../../stores/store';
import UserProfile from '../../../../utils/auth';
import { isLoading } from '../../../../utils/shared';
import { useAqueductConsts } from '../../../hooks/useAqueductConsts';
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

  const { apiAddress } = useAqueductConsts();

  // Set the title of the page on page load.
  useEffect(() => {
    document.title = 'Artifact | Aqueduct';
  }, []);

  // TODO: Fetch artifact data and render here.
  useEffect(() => {
    console.log('Fetching artifact data ...');
    console.log('Url params: ');
    console.log('workflowId: ', workflowId);
    console.log('workflowDagResultId: ', workflowDagResultId);
    console.log('artifactId: ', artifactId);
    console.log('workflow regular useEffect: ', workflow);
    //console.log('artifactResult: ', artifactResult);

    // Fetching the workflow by Id:
    // TODO: Might not need this call after all.
    //dispatch(handleGetWorkflow({ apiKey: user.apiKey, workflowId }));

    console.log('fetching the artifact Result');
    dispatch(
      handleGetArtifactResults({
        apiKey: user.apiKey,
        workflowDagResultId,
        artifactId,
      })
    );
  }, []);

  // After artifact details are fetched, get the workflow details
  useEffect(() => {
    // Fetch workflow details
    console.log('other useEffect workflow: ', workflow);
    // only get workflow if it's not currently loading one.
    const loadingStatus = workflow.loadingStatus;

    if (Object.keys(workflow.dags).length < 1 && !isLoading(loadingStatus)) {
      console.log('Fetching workflow inside if statement ...');
      dispatch(handleGetWorkflow({ apiKey: user.apiKey, workflowId }));
    }
  }, [artifactResult, workflowId]);

  const artifactMetadata =
    workflow?.dags[workflowDagResultId]?.artifacts[artifactId];
  console.log('artifactMetadata: ', artifactMetadata);

  if (!artifactResult || !artifactResult.result) {
    return (
      <Layout user={user}>
        <CircularProgress />
      </Layout>
    );
  }

  const parsedData = JSON.parse(artifactResult.result.data);
  console.log('artifact details parsedData: ', parsedData);

  return (
    <Layout user={user}>
      <Box width={'800px'}>
        <Box width="100%">
          <Box width="100%" display="flex">
            <ArtifactDetailsHeader artifactName="churn_table" />
            <Button variant="contained" sx={{ maxHeight: '32px' }}>
              EXPORT
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
              <KeyValueTable
                rows={mockMetrics}
                tableType={KeyValueTableType.Metric}
              />
            </Box>
            <Box width="96px" />
            <Box width="100%">
              <Typography variant="h5" component="div" marginBottom="8px">
                Checks
              </Typography>
              <KeyValueTable
                rows={mockChecks}
                tableType={KeyValueTableType.Check}
              />
            </Box>
          </Box>
        </Box>
      </Box>
    </Layout>
  );
};

export default ArtifactDetailsPage;
