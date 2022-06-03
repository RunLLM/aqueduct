import useUser from '@aqueducthq/common/src/components/hooks/useUser';
import HomePage from '@aqueducthq/common/src/components/pages/HomePage';
import { useRouter } from 'next/router';
import React from 'react';

export { getServerSideProps } from '@aqueducthq/common/src/components/pages/getServerSideProps';

const Home: React.FC = () => {
    const router = useRouter();
    const { user, loading, success } = useUser();

    if (loading) {
        return null;
    }

    if (!success) {
        router.push('/login');
        return null;
    }

    return <HomePage user={user} />;
};

export default Home;
