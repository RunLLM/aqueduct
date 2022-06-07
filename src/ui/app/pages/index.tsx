import { HomePage, useUser } from '@aqueducthq/common';
import { useRouter } from 'next/router';
import React from 'react';
export { getServerSideProps } from '@aqueducthq/common';

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

    return <HomePage user={user} />;
};

export default Home;
