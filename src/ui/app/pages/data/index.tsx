import { DataPage, useUser } from '@aqueducthq/common';
import { useRouter } from 'next/router';
import React from 'react';
export { getServerSideProps } from '@aqueducthq/common';

const Data: React.FC = () => {
    const router = useRouter();
    const { user, loading, success } = useUser();
    if (loading) {
        return null;
    }

    if (!user || !success) {
        router.push('/login');
        return null;
    }

    return <DataPage user={user} />;
};

export default Data;
