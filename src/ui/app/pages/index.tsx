import useUser from '@aqueducthq/common/src/components/hooks/useUser';
import HomePage from '@aqueducthq/common/src/components/pages/HomePage';
import { useRouter } from 'next/router';
import React from 'react';

export { getServerSideProps } from '@aqueducthq/common/src/components/pages/getServerSideProps';

const Home: React.FC = () => {
    const { user, loading, success } = useUser();
    if (loading) {
        return null;
    }

    if (!success) {
        const router = useRouter();
        router.push('/login');
        return null;
    }

    //return <GettingStartedTutorial user={user} />;

    return <HomePage user={user} />;
};

export default Home;
