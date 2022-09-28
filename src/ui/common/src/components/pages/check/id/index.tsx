import { faChevronRight } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { CircularProgress, Link, List, ListItem } from '@mui/material';
import Accordion from '@mui/material/Accordion';
import AccordionDetails from '@mui/material/AccordionDetails';
import AccordionSummary from '@mui/material/AccordionSummary';
import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React, { useEffect, useState } from 'react';
import Plot from 'react-plotly.js';
import { useDispatch, useSelector } from 'react-redux';
import { Link as RouterLink, useParams } from 'react-router-dom';

import PaginatedTable from '../../../../components/tables/PaginatedTable';
import { artifactTypeToIconMapping } from '../../../../components/workflows/nodes/nodeTypes';
import { handleGetWorkflowDagResult } from '../../../../handlers/getWorkflowDagResult';
import { handleListArtifactResults } from '../../../../handlers/listArtifactResults';
import { AppDispatch, RootState } from '../../../../stores/store';
import UserProfile from '../../../../utils/auth';
import { Data } from '../../../../utils/data';
import { getPathPrefix } from '../../../../utils/getPathPrefix';
import { isFailed, isInitial, isLoading } from '../../../../utils/shared';
import DefaultLayout from '../../../layouts/default';
import DetailsPageHeader from '../../components/DetailsPageHeader';
import { LayoutProps } from '../../types';

type CheckDetailsPageProps = {
    user: UserProfile;
    Layout?: React.FC<LayoutProps>;
};

const CheckDetailsPage: React.FC<CheckDetailsPageProps> = ({
    user,
    Layout = DefaultLayout,
}) => {
    const dispatch: AppDispatch = useDispatch();
    const { workflowId, workflowDagResultId, checkOperatorId } = useParams();

    const [inputsExpanded, setInputsExpanded] = useState<boolean>(true);
    const [outputsExpanded, setOutputsExpanded] = useState<boolean>(true);

    const workflowDagResultWithLoadingStatus = useSelector(
        (state: RootState) =>
            state.workflowDagResultsReducer.results[workflowDagResultId]
    );

    const operator = (workflowDagResultWithLoadingStatus?.result?.operators ??
        {})[checkOperatorId];

    const artifactId = operator?.outputs[0];

    const artifactHistoryWithLoadingStatus = useSelector((state: RootState) =>
        !!artifactId
            ? state.artifactResultsReducer.artifacts[artifactId]
            : undefined
    );

    console.log('workflowDagResultWithLoadingStatus: ', workflowDagResultWithLoadingStatus);

    console.log('workflowId: ', workflowId);
    console.log('workflowDagResultId: ', workflowDagResultId);
    console.log('checkOperatorId: ', checkOperatorId);

    useEffect(() => {
        document.title = 'Check Details | Aqueduct';

        // Load workflow dag result if it's not cached
        if (
            !workflowDagResultWithLoadingStatus ||
            isInitial(workflowDagResultWithLoadingStatus.status)
        ) {
            console.log('Loading workflowDagResult');
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
        // Load artifact history once workflow dag results finished loading
        // and the result is not cached
        if (
            !artifactHistoryWithLoadingStatus &&
            !!artifactId &&
            !!workflowDagResultWithLoadingStatus &&
            !isInitial(workflowDagResultWithLoadingStatus.status) &&
            !isLoading(workflowDagResultWithLoadingStatus.status)
        ) {
            dispatch(
                handleListArtifactResults({
                    apiKey: user.apiKey,
                    workflowId,
                    artifactId,
                })
            );
        }
    }, [workflowDagResultWithLoadingStatus, artifactId]);

    console.log('operator: ', operator);

    useEffect(() => {
        if (!!operator) {
            document.title = `${operator.name} | Aqueduct`;
        }
    }, [operator]);

    const listStyle = {
        width: '100%',
        maxWidth: 360,
        bgcolor: 'background.paper',
    };

    if (
        !workflowDagResultWithLoadingStatus ||
        isInitial(workflowDagResultWithLoadingStatus.status) ||
        isLoading(workflowDagResultWithLoadingStatus.status)
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

    const mockCheckArtifactData = [
        // severity in the mock corresponnds with level in the API response.
        { status: 'succeeded', level: 'warning', result: 'True', date_completed: '3/14/2022 4:00 PST' },
        { status: 'succeeded', level: 'warning', result: 'False', date_completed: '3/14/2022 4:00 PST' },
        { status: 'succeeded', level: 'error', result: 'True', date_completed: '3/14/2022 4:00 PST' }
    ]
    const historicalCheckData: Data = {
        schema: {
            fields: [
                { name: 'status', type: 'varchar' },
                { name: 'level', type: 'varchar' },
                { name: 'result', type: 'varchar' },
                { name: 'date_completed', type: 'varchar' }
            ],
            pandas_version: '0.0.1', // Not sure what actual value to put here, just filling in for now :)
        },
        data: mockCheckArtifactData
    };

    // Function to get the numerical value of the metric output
    // TODO: Use this inside of the accordion component below.
    // NOTE: This code is shared with the metric details page, perhaps we should make this into a hook or component.
    const operatorOutputsList = operator.outputs.map((artifactId) => {
        const artifactResult = (workflowDagResultWithLoadingStatus.result
            ?.artifacts ?? {})[artifactId];
        if (!artifactResult) {
            return null;
        }

        if (
            !artifactResult.result ||
            artifactResult.result.content_serialized === undefined
        ) {
            // Link to appropriate artifact details page
            // Show tableIcon here as part of the link.
            return (
                <Box key={artifactId}>
                    <Link
                        to={`${getPathPrefix()}/workflow/${workflowId}/result/${workflowDagResultId}/artifact/${artifactId}`}
                        component={RouterLink as any}
                        sx={{ marginLeft: '16px' }}
                        underline="none"
                    >
                        {artifactResult.name}
                    </Link>
                </Box>
            );
        }

        return (
            <Box key={artifactId}>
                <Typography variant="body1">
                    {artifactResult.result.content_serialized}
                </Typography>
            </Box>
        );
    });

    return (
        <Layout user={user}>
            <Box width={'800px'}>
                <Box width="100%">
                    <Box width="100%">
                        <DetailsPageHeader name={operator?.name} />
                        {operator?.description && (
                            <Typography variant="body1">{operator.description}</Typography>
                        )}
                    </Box>
                </Box>

                <Box width="100%" marginTop="32px">
                    <Typography variant="h5" marginBottom="8px">Recent Results</Typography>
                    <PaginatedTable data={historicalCheckData} />
                </Box>

                <Box width="100%" marginTop="32px">
                    <Typography variant="h5">Related Outputs</Typography>
                </Box>

            </Box>
        </Layout>
    );
};

export default CheckDetailsPage;
