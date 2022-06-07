import { HomePage, useUser } from '@aqueducthq/common';
import { useRouter } from 'next/router';
import React from 'react';
export { getServerSideProps } from '@aqueducthq/common';

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
