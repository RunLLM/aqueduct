import Box from '@mui/material/Box';
import React, { useEffect } from 'react';

import { BreadcrumbLink } from '../../components/layouts/NavBar';
import UserProfile from '../../utils/auth';
import GettingStartedTutorial from '../cards/GettingStartedTutorial';
import DefaultLayout, { DefaultLayoutMargin } from '../layouts/default';
import { LayoutProps } from './types';

type HomePageProps = {
  user: UserProfile;
  Layout?: React.FC<LayoutProps>;
};

const HomePage: React.FC<HomePageProps> = ({
  user,
  Layout = DefaultLayout,
}) => {
  // Set the title of the page on page load.
  useEffect(() => {
    document.title = 'Home | Aqueduct';
  }, []);

  return (
    <Layout breadcrumbs={[BreadcrumbLink.HOME]} user={user}>
      <Box paddingBottom={DefaultLayoutMargin}>
        <GettingStartedTutorial user={user} />
      </Box>
    </Layout>
  );
};

export default HomePage;
