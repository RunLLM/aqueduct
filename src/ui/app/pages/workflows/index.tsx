import { useUser, WorkflowsPage } from '@aqueducthq/common';
import { useRouter } from 'next/router';
import React from 'react';
export { getServerSideProps } from '@aqueducthq/common';

const Workflows: React.FC = () => {
    const router = useRouter();
    const { user, loading, success } = useUser();

    if (loading) {
        return null;
    }

    if (!success) {
        router.push('/login');
        return null;
    }

    return <WorkflowsPage user={user} />;
};

export default Workflows;
