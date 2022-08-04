import Box from '@mui/material/Box';
import Link from '@mui/material/Link';
import Typography from '@mui/material/Typography';
import React, { useEffect } from 'react';
import StickyHeaderTable from '../../../tables/StickyHeaderTable';

import UserProfile from '../../../../utils/auth';
import { useAqueductConsts } from '../../../hooks/useAqueductConsts';
import DefaultLayout from '../../../layouts/default';
import { LayoutProps } from '../../types';
import { Button } from '@mui/material';

type ArtifactDetailsHeaderProps = {
    artifactName: string;
    lastUpdated: string;
    sourceLocation: string;
}

const ArtifactDetailsHeader: React.FC<ArtifactDetailsHeaderProps> = ({ artifactName, lastUpdated, sourceLocation }) => {
    return (
        <Box width="100%">
            <Typography variant="h4" component="div">
                {artifactName}
            </Typography>
            <Typography marginTop="4px" variant="caption" component="div">Last Updated: {lastUpdated}</Typography>
            <Typography variant="caption" component="div">Source: <Link>{sourceLocation}</Link></Typography>
        </Box>
    )
}

type ArtifactDetailsPageProps = {
    user: UserProfile;
    Layout?: React.FC<LayoutProps>;
};

const ArtifactDetailsPage: React.FC<ArtifactDetailsPageProps> = ({
    user,
    Layout = DefaultLayout,
}) => {
    // Set the title of the page on page load.
    useEffect(() => {
        document.title = 'Artifact | Aqueduct';
    }, []);

    const { apiAddress } = useAqueductConsts();
    return (
        <Layout user={user}>
            <Box width={'800px'}>
                <Box width="100%">
                    <Box width="100%" display='flex'>
                        <ArtifactDetailsHeader artifactName="churn_table" lastUpdated="3/17/2022 10:00pm" sourceLocation="s3/myBucket" />
                        <Button variant="contained" sx={{ maxHeight: '36px' }}>EXPORT</Button>
                    </Box>
                    <Box width="100%" marginTop="12px">
                        <Typography variant="h5" component="div" marginBottom="8px">Preview</Typography>
                        <StickyHeaderTable />
                    </Box>
                </Box>
            </Box>
        </Layout>
    );
};

export default ArtifactDetailsPage;
