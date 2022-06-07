import { DataPage, useUser } from '@aqueducthq/common';
import { useRouter } from 'next/router';
import React from 'react';
export { getServerSideProps } from '@aqueducthq/common';

const Data: React.FC = () => {
    const { user, loading, success } = useUser();
    if (loading) {
        return null;
    }

    const router = useRouter();
    if (!success) {
        router.push('/login');
        return null;
    }

    return <DataPage user={user} />;
};

export default Data;
