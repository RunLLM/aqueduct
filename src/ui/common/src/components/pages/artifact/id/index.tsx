import Typography from '@mui/material/Typography';
import React, { useEffect } from 'react';

import UserProfile from '../../../../utils/auth';
import { useAqueductConsts } from '../../../hooks/useAqueductConsts';
import DefaultLayout from '../../../layouts/default';
import { LayoutProps } from '../../types';

type AccountPageProps = {
  user: UserProfile;
  Layout?: React.FC<LayoutProps>;
};

const ArtifactDetailsPage: React.FC<AccountPageProps> = ({
  user,
  Layout = DefaultLayout,
}) => {
  // Set the title of the page on page load.
  useEffect(() => {
    document.title = 'Account | Aqueduct';
  }, []);

  const { apiAddress } = useAqueductConsts();

  // TODO: Implement new header component here.
  return (
    <Layout user={user}>
      <Typography variant="h2" gutterBottom component="div">
        Artifact Details
      </Typography>
    </Layout>
  );
};

export default ArtifactDetailsPage;
