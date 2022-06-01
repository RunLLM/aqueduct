import useUser from '@aqueducthq/common/src/components/hooks/useUser';
import WorkflowPage from '@aqueducthq/common/src/components/pages/workflow/id';
import { useRouter } from 'next/router';
import React from 'react';
export { getServerSideProps } from '@aqueducthq/common/src/components/pages/getServerSideProps';

const Workflow: React.FC = () => {
    const { user, loading, success } = useUser();
    if (loading) {
        return null;
    }

    const router = useRouter();
    if (!success) {
        router.push('/login');
        return null;
    }

    const workflowId = router.query.id as string;
    return <WorkflowPage user={user} workflowId={workflowId} />;
};

export default Workflow;
