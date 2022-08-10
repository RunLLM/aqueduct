import { faCircleCheck } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Button, CircularProgress } from '@mui/material';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React, { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useParams } from 'react-router-dom';
import { Data, DataSchema } from 'src/utils/data';
import React, { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useNavigate, useParams } from 'react-router-dom';

import {
    ArtifactResult,
    handleGetArtifactResults,
} from '../../../../reducers/workflow';
import { AppDispatch, RootState } from '../../../../stores/store';
import UserProfile from '../../../../utils/auth';
import { exportCsv } from '../../../../utils/preview';
import DefaultLayout from '../../../layouts/default';
import StickyHeaderTable from '../../../tables/StickyHeaderTable';
import KeyValueTable from '../../../tables/KeyValueTable';
import StickyHeaderTable from '../../../tables/StickyHeaderTable';
import { LayoutProps } from '../../types';
import { Button, CircularProgress } from '@mui/material';
import { useNavigate, useParams } from 'react-router-dom';
import { ArtifactResult, handleGetArtifactResults, handleGetWorkflow } from '../../../../reducers/workflow';
import { useDispatch, useSelector } from 'react-redux';
import { AppDispatch, RootState } from '../../../../stores/store';

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

    useEffect(() => {
        console.log('workflow workflowUseEffect: ', workflow);
        //console.log('artifactResult: ', artifactResult);
    }, [workflow]);

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
                        <ArtifactDetailsHeader
                            artifactName="churn_table"
                            lastUpdated="3/17/2022 10:00pm"
                            sourceLocation="s3/myBucket"
                        />
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
                            <KeyValueTable />
                        </Box>
                        <Box width="96px" />
                        <Box width="100%">
                            <Typography variant="h5" component="div" marginBottom="8px">
                                Checks
                            </Typography>
                            <KeyValueTable />
                        </Box>
                    </Box>
                </Box>
            </Box>
        </Layout>
    );
};

export default ArtifactDetailsPage;
