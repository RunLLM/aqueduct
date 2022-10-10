import React, { useEffect } from 'react';

import { BreadcrumbLinks } from '../../components/layouts/NavBar';
import UserProfile from '../../utils/auth';
import GettingStartedTutorial from '../cards/GettingStartedTutorial';
import DefaultLayout from '../layouts/default';
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
    <Layout breadcrumbs={[BreadcrumbLinks.HOME]} user={user}>
      <div />
      <GettingStartedTutorial user={user} />
    </Layout>
  );
};

export default HomePage;
