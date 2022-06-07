import { useUser, WorkflowPage } from '@aqueducthq/common';
import { useRouter } from 'next/router';
import React from 'react';
export { getServerSideProps } from '@aqueducthq/common';

const Workflow: React.FC = () => {
    const router = useRouter();
    const workflowId = router.query.id as string;
    const { user, loading, success } = useUser();

    if (loading) {
        return null;
    }

    if (!user || !success) {
        router.push('/login');
        return null;
    }

    return <WorkflowPage user={user} workflowId={workflowId} />;
};

export default Workflow;
