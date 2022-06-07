import useUser from '@aqueducthq/common/src/components/hooks/useUser';
import WorkflowPage from '@aqueducthq/common/src/components/pages/workflow/id';
import { useRouter } from 'next/router';
import React from 'react';
export { getServerSideProps } from '@aqueducthq/common/src/components/pages/getServerSideProps';

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
