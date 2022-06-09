import React, { useEffect } from 'react';

import UserProfile from '../../utils/auth';
import GettingStartedTutorial from '../cards/GettingStartedTutorial';
import DefaultLayout from '../layouts/default';

type HomePageProps = {
  user: UserProfile;
};

const HomePage: React.FC<HomePageProps> = ({ user }) => {
  // Set the title of the page on page load.
  useEffect(() => {
    document.title = 'Home | Aqueduct';
  }, []);

  return (
    <DefaultLayout user={user}>
      <div />
      <GettingStartedTutorial user={user} />
    </DefaultLayout>
  );
};

export default HomePage;
